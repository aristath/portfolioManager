"""Balance service - atomic virtual cash operations."""

import logging
from typing import Dict, List, Optional

from app.modules.satellites.database.balance_repository import BalanceRepository
from app.modules.satellites.database.bucket_repository import BucketRepository
from app.modules.satellites.domain.enums import BucketStatus, TransactionType
from app.modules.satellites.domain.models import BucketBalance, BucketTransaction

logger = logging.getLogger(__name__)


class BalanceService:
    """Service for atomic virtual cash operations.

    All operations that affect cash balances are atomic - they update
    the balance and record a transaction in a single database transaction.

    This ensures the critical invariant:
    SUM(bucket_balances for currency X) == Actual brokerage balance for currency X
    """

    def __init__(
        self,
        balance_repo: Optional[BalanceRepository] = None,
        bucket_repo: Optional[BucketRepository] = None,
    ):
        self._balance_repo = balance_repo or BalanceRepository()
        self._bucket_repo = bucket_repo or BucketRepository()

    # Query methods

    async def get_balance(
        self, bucket_id: str, currency: str
    ) -> Optional[BucketBalance]:
        """Get balance for a bucket in a specific currency."""
        return await self._balance_repo.get_balance(bucket_id, currency)

    async def get_balance_amount(self, bucket_id: str, currency: str) -> float:
        """Get balance amount, returning 0 if not found."""
        return await self._balance_repo.get_balance_amount(bucket_id, currency)

    async def get_all_balances(self, bucket_id: str) -> List[BucketBalance]:
        """Get all currency balances for a bucket."""
        return await self._balance_repo.get_all_balances(bucket_id)

    async def get_total_by_currency(self, currency: str) -> float:
        """Get total virtual balance across all buckets for a currency."""
        return await self._balance_repo.get_total_by_currency(currency)

    async def get_portfolio_summary(self) -> Dict[str, Dict[str, float]]:
        """Get summary of all buckets and their balances.

        Returns:
            Dict mapping bucket_id to dict of currency -> balance
        """
        buckets = await self._bucket_repo.get_all()
        summary = {}

        for bucket in buckets:
            balances = await self._balance_repo.get_all_balances(bucket.id)
            summary[bucket.id] = {b.currency: b.balance for b in balances}

        return summary

    # Atomic cash operations

    async def record_trade_settlement(
        self,
        bucket_id: str,
        amount: float,
        currency: str,
        is_buy: bool,
        description: Optional[str] = None,
    ) -> BucketBalance:
        """Record a trade settlement, atomically updating balance.

        For buys: subtracts amount from balance (cash goes out)
        For sells: adds amount to balance (cash comes in)

        Args:
            bucket_id: ID of the bucket
            amount: Absolute trade amount (always positive)
            currency: Currency code
            is_buy: True for buy (cash out), False for sell (cash in)
            description: Optional description for audit trail

        Returns:
            Updated balance
        """
        if amount < 0:
            raise ValueError("Amount must be positive")

        # For buys, cash goes out (negative delta)
        # For sells, cash comes in (positive delta)
        delta = -amount if is_buy else amount
        tx_type = TransactionType.TRADE_BUY if is_buy else TransactionType.TRADE_SELL

        # Adjust balance
        balance = await self._balance_repo.adjust_balance(bucket_id, currency, delta)

        # Record transaction
        tx = BucketTransaction(
            bucket_id=bucket_id,
            type=tx_type,
            amount=amount if is_buy else amount,  # Store as positive
            currency=currency.upper(),
            description=description or f"{'Buy' if is_buy else 'Sell'} settlement",
        )
        await self._balance_repo.record_transaction(tx)

        logger.info(
            f"Recorded trade settlement for '{bucket_id}': "
            f"{'bought' if is_buy else 'sold'} {amount:.2f} {currency}"
        )
        return balance

    async def record_dividend(
        self,
        bucket_id: str,
        amount: float,
        currency: str,
        description: Optional[str] = None,
    ) -> BucketBalance:
        """Record a dividend payment.

        Args:
            bucket_id: ID of the bucket receiving the dividend
            amount: Dividend amount (positive)
            currency: Currency code
            description: Optional description

        Returns:
            Updated balance
        """
        if amount <= 0:
            raise ValueError("Dividend amount must be positive")

        # Adjust balance (cash comes in)
        balance = await self._balance_repo.adjust_balance(bucket_id, currency, amount)

        # Record transaction
        tx = BucketTransaction(
            bucket_id=bucket_id,
            type=TransactionType.DIVIDEND,
            amount=amount,
            currency=currency.upper(),
            description=description or "Dividend received",
        )
        await self._balance_repo.record_transaction(tx)

        logger.info(f"Recorded dividend for '{bucket_id}': {amount:.2f} {currency}")
        return balance

    async def transfer_between_buckets(
        self,
        from_bucket_id: str,
        to_bucket_id: str,
        amount: float,
        currency: str,
        description: Optional[str] = None,
    ) -> tuple[BucketBalance, BucketBalance]:
        """Transfer cash between buckets.

        Args:
            from_bucket_id: Source bucket ID
            to_bucket_id: Destination bucket ID
            amount: Amount to transfer (positive)
            currency: Currency code
            description: Optional description

        Returns:
            Tuple of (from_balance, to_balance) after transfer

        Raises:
            ValueError: If insufficient funds or invalid buckets
        """
        if amount <= 0:
            raise ValueError("Transfer amount must be positive")

        if from_bucket_id == to_bucket_id:
            raise ValueError("Cannot transfer to same bucket")

        # Validate buckets exist
        from_bucket = await self._bucket_repo.get_by_id(from_bucket_id)
        to_bucket = await self._bucket_repo.get_by_id(to_bucket_id)

        if not from_bucket:
            raise ValueError(f"Source bucket '{from_bucket_id}' not found")
        if not to_bucket:
            raise ValueError(f"Destination bucket '{to_bucket_id}' not found")

        # Check sufficient funds
        current_balance = await self._balance_repo.get_balance_amount(
            from_bucket_id, currency
        )
        if current_balance < amount:
            raise ValueError(
                f"Insufficient funds in '{from_bucket_id}': "
                f"has {current_balance:.2f}, needs {amount:.2f} {currency}"
            )

        # Check core minimum (if transferring from core)
        if from_bucket_id == "core":
            settings = await self._balance_repo.get_all_allocation_settings()
            satellite_budget_pct = settings.get("satellite_budget_pct", 0.0)
            core_min_pct = 1.0 - satellite_budget_pct

            # Get total portfolio value for validation
            total = await self.get_total_by_currency(currency)
            if total > 0:
                remaining_after_transfer = current_balance - amount
                remaining_pct = remaining_after_transfer / total
                if remaining_pct < core_min_pct:
                    raise ValueError(
                        f"Transfer would put core below minimum allocation "
                        f"({remaining_pct:.1%} < {core_min_pct:.1%})"
                    )

        # Perform the transfer
        from_balance = await self._balance_repo.adjust_balance(
            from_bucket_id, currency, -amount
        )
        to_balance = await self._balance_repo.adjust_balance(
            to_bucket_id, currency, amount
        )

        # Record transactions for audit trail
        desc = description or f"Transfer from {from_bucket_id} to {to_bucket_id}"

        tx_out = BucketTransaction(
            bucket_id=from_bucket_id,
            type=TransactionType.TRANSFER_OUT,
            amount=amount,
            currency=currency.upper(),
            description=desc,
        )
        tx_in = BucketTransaction(
            bucket_id=to_bucket_id,
            type=TransactionType.TRANSFER_IN,
            amount=amount,
            currency=currency.upper(),
            description=desc,
        )
        await self._balance_repo.record_transaction(tx_out)
        await self._balance_repo.record_transaction(tx_in)

        logger.info(
            f"Transferred {amount:.2f} {currency} "
            f"from '{from_bucket_id}' to '{to_bucket_id}'"
        )
        return from_balance, to_balance

    async def allocate_deposit(
        self,
        total_amount: float,
        currency: str,
        description: Optional[str] = None,
    ) -> Dict[str, float]:
        """Allocate a new deposit across buckets based on targets.

        Deposits are split among buckets that are below their target allocation.
        Priority goes to buckets furthest below target.

        Args:
            total_amount: Total deposit amount
            currency: Currency code
            description: Optional description

        Returns:
            Dict mapping bucket_id to allocated amount
        """
        if total_amount <= 0:
            raise ValueError("Deposit amount must be positive")

        # Get allocation settings
        settings = await self._balance_repo.get_all_allocation_settings()
        satellite_budget_pct = settings.get("satellite_budget_pct", 0.0)

        # Get all active buckets
        buckets = await self._bucket_repo.get_active()

        # Calculate current total and target amounts
        current_total = await self.get_total_by_currency(currency)
        new_total = current_total + total_amount

        allocations: Dict[str, float] = {}
        remaining = total_amount

        # Core always gets at least its target
        core_target_pct = 1.0 - satellite_budget_pct
        core_target_amount = new_total * core_target_pct
        core_current = await self._balance_repo.get_balance_amount("core", currency)
        core_needed = max(0, core_target_amount - core_current)

        if core_needed > 0:
            core_allocation = min(core_needed, remaining)
            allocations["core"] = core_allocation
            remaining -= core_allocation

        # Distribute remaining to satellites below target
        if remaining > 0:
            satellites = [b for b in buckets if b.type.value == "satellite"]

            # Filter to only accumulating or active satellites
            eligible = [
                s
                for s in satellites
                if s.status in (BucketStatus.ACCUMULATING, BucketStatus.ACTIVE)
                and s.target_pct
                and s.target_pct > 0
            ]

            if eligible:
                # Calculate how far below target each is
                deficits = []
                for sat in eligible:
                    # target_pct is guaranteed non-None by the filter above
                    target_amount = new_total * (sat.target_pct or 0.0)
                    current = await self._balance_repo.get_balance_amount(
                        sat.id, currency
                    )
                    deficit = max(0, target_amount - current)
                    if deficit > 0:
                        deficits.append((sat.id, deficit))

                # Distribute proportionally to deficits
                total_deficit = sum(d[1] for d in deficits)
                if total_deficit > 0:
                    for sat_id, deficit in deficits:
                        share = (deficit / total_deficit) * remaining
                        allocations[sat_id] = share

        # Record the allocations
        for bucket_id, amount in allocations.items():
            if amount > 0:
                await self._balance_repo.adjust_balance(bucket_id, currency, amount)

                tx = BucketTransaction(
                    bucket_id=bucket_id,
                    type=TransactionType.DEPOSIT,
                    amount=amount,
                    currency=currency.upper(),
                    description=description or "Deposit allocation",
                )
                await self._balance_repo.record_transaction(tx)

        logger.info(
            f"Allocated deposit of {total_amount:.2f} {currency}: {allocations}"
        )
        return allocations

    async def reallocate(
        self,
        from_bucket_id: str,
        to_bucket_id: str,
        amount: float,
        currency: str,
    ) -> tuple[BucketBalance, BucketBalance]:
        """Reallocate funds between buckets (meta-allocator operation).

        Similar to transfer but uses REALLOCATION transaction type
        to distinguish quarterly reallocation from manual transfers.

        Args:
            from_bucket_id: Source bucket ID
            to_bucket_id: Destination bucket ID
            amount: Amount to reallocate
            currency: Currency code

        Returns:
            Tuple of (from_balance, to_balance)
        """
        if amount <= 0:
            raise ValueError("Reallocation amount must be positive")

        # Adjust balances
        from_balance = await self._balance_repo.adjust_balance(
            from_bucket_id, currency, -amount
        )
        to_balance = await self._balance_repo.adjust_balance(
            to_bucket_id, currency, amount
        )

        # Record as reallocation (not regular transfer)
        tx_out = BucketTransaction(
            bucket_id=from_bucket_id,
            type=TransactionType.REALLOCATION,
            amount=-amount,  # Negative for outflow
            currency=currency.upper(),
            description=f"Quarterly reallocation to {to_bucket_id}",
        )
        tx_in = BucketTransaction(
            bucket_id=to_bucket_id,
            type=TransactionType.REALLOCATION,
            amount=amount,  # Positive for inflow
            currency=currency.upper(),
            description=f"Quarterly reallocation from {from_bucket_id}",
        )
        await self._balance_repo.record_transaction(tx_out)
        await self._balance_repo.record_transaction(tx_in)

        logger.info(
            f"Reallocated {amount:.2f} {currency} "
            f"from '{from_bucket_id}' to '{to_bucket_id}'"
        )
        return from_balance, to_balance

    # Transaction history

    async def get_transactions(
        self,
        bucket_id: str,
        limit: int = 100,
        transaction_type: Optional[TransactionType] = None,
    ) -> List[BucketTransaction]:
        """Get transaction history for a bucket."""
        return await self._balance_repo.get_transactions(
            bucket_id, limit=limit, transaction_type=transaction_type
        )

    async def get_recent_transactions(
        self, bucket_id: str, days: int = 30
    ) -> List[BucketTransaction]:
        """Get recent transactions for a bucket."""
        return await self._balance_repo.get_recent_transactions(bucket_id, days)

    # Settings

    async def get_allocation_settings(self) -> Dict[str, float]:
        """Get all allocation settings."""
        return await self._balance_repo.get_all_allocation_settings()

    async def update_satellite_budget(self, budget_pct: float) -> None:
        """Update the global satellite budget percentage.

        Args:
            budget_pct: New budget percentage (0.0-1.0)

        Raises:
            ValueError: If budget is out of range
        """
        if not 0.0 <= budget_pct <= 0.30:  # Max 30% for satellites
            raise ValueError("Satellite budget must be between 0% and 30%")

        await self._balance_repo.set_allocation_setting(
            "satellite_budget_pct", budget_pct
        )
        logger.info(f"Updated satellite budget to {budget_pct:.1%}")
