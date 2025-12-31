"""Deposit processor for automatic bucket allocation.

Integrates with cash flow sync to automatically split deposits across buckets
based on target allocations.
"""

import logging
from typing import Optional

logger = logging.getLogger(__name__)


async def process_deposit(
    amount: float,
    currency: str,
    transaction_id: Optional[str] = None,
    description: Optional[str] = None,
) -> dict:
    """Process a deposit by allocating it across buckets.

    This function is called when a new deposit cash flow is detected.
    It uses the BalanceService to split the deposit based on target allocations.

    Args:
        amount: Deposit amount
        currency: Currency code (EUR, USD, etc.)
        transaction_id: Optional transaction ID from cash flow
        description: Optional description

    Returns:
        Dict with allocation results:
        {
            "total_amount": float,
            "currency": str,
            "allocations": {"bucket_id": amount, ...},
            "transaction_id": str or None,
        }
    """
    # Check if satellites module is available
    try:
        from app.modules.satellites.services.balance_service import BalanceService
    except ImportError:
        logger.debug(
            "Satellites module not available - deposit goes to core bucket only"
        )
        return {
            "total_amount": amount,
            "currency": currency,
            "allocations": {"core": amount},
            "transaction_id": transaction_id,
        }

    try:
        balance_service = BalanceService()

        # Allocate deposit across buckets
        allocations = await balance_service.allocate_deposit(
            total_amount=amount,
            currency=currency,
            description=description
            or f"Deposit allocation (txn: {transaction_id or 'N/A'})",
        )

        logger.info(
            f"Processed deposit: {amount:.2f} {currency} → "
            f"{len(allocations)} buckets (txn: {transaction_id or 'N/A'})"
        )

        for bucket_id, allocated_amount in allocations.items():
            logger.info(f"  - {bucket_id}: {allocated_amount:.2f} {currency}")

        return {
            "total_amount": amount,
            "currency": currency,
            "allocations": allocations,
            "transaction_id": transaction_id,
        }

    except Exception as e:
        logger.error(f"Failed to process deposit allocation: {e}", exc_info=True)
        # Fall back to core bucket if allocation fails
        return {
            "total_amount": amount,
            "currency": currency,
            "allocations": {"core": amount},
            "transaction_id": transaction_id,
            "error": str(e),
        }


async def should_process_cash_flow(cash_flow: dict) -> bool:
    """Determine if a cash flow should trigger deposit processing.

    Args:
        cash_flow: Cash flow record dict

    Returns:
        True if this is a deposit that should be allocated across buckets
    """
    transaction_type = (
        cash_flow.get("type") or cash_flow.get("transaction_type", "")
    ).lower()

    # Process these types as deposits:
    # - "deposit" - explicit deposit
    # - "refill" - account funding
    # - "transfer_in" - incoming transfer
    deposit_types = {"deposit", "refill", "transfer_in"}

    return transaction_type in deposit_types
