"""Unit tests for BalanceService."""

from unittest.mock import AsyncMock, MagicMock

import pytest

from app.modules.satellites.domain.enums import (
    BucketStatus,
    BucketType,
    TransactionType,
)
from app.modules.satellites.domain.models import (
    Bucket,
    BucketBalance,
    BucketTransaction,
)
from app.modules.satellites.services.balance_service import BalanceService


@pytest.fixture
def mock_balance_repo():
    """Create a mock balance repository."""
    repo = MagicMock()
    repo.get_balance = AsyncMock()
    repo.get_balance_amount = AsyncMock(return_value=0.0)
    repo.set_balance = AsyncMock()
    repo.adjust_balance = AsyncMock()
    repo.record_transaction = AsyncMock()
    repo.get_all_allocation_settings = AsyncMock(
        return_value={
            "satellite_budget_pct": 0.30,
            "satellite_min_pct": 0.02,
            "satellite_max_pct": 0.15,
        }
    )
    repo.get_allocation_setting = AsyncMock(return_value=0.30)
    repo.get_total_by_currency = AsyncMock(return_value=0.0)
    return repo


@pytest.fixture
def mock_bucket_repo():
    """Create a mock bucket repository."""
    repo = MagicMock()
    repo.get_all = AsyncMock()
    repo.get_by_id = AsyncMock()
    repo.get_core = AsyncMock()
    return repo


@pytest.fixture
def balance_service(mock_balance_repo, mock_bucket_repo):
    """Create a BalanceService with mocked dependencies."""
    return BalanceService(
        balance_repo=mock_balance_repo,
        bucket_repo=mock_bucket_repo,
    )


class TestBalanceService:
    """Tests for BalanceService."""

    @pytest.mark.asyncio
    async def test_record_trade_settlement_buy(
        self, balance_service, mock_balance_repo, mock_bucket_repo
    ):
        """Test recording a buy trade settlement."""
        mock_bucket_repo.get_by_id.return_value = Bucket(
            id="core",
            name="Core",
            type=BucketType.CORE,
            status=BucketStatus.ACTIVE,
        )
        mock_balance_repo.adjust_balance.return_value = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=500.0,
            last_updated="2024-01-15",
        )
        mock_balance_repo.record_transaction.return_value = BucketTransaction(
            id=1,
            bucket_id="core",
            type=TransactionType.TRADE_BUY,
            amount=-500.0,
            currency="EUR",
        )

        balance = await balance_service.record_trade_settlement(
            bucket_id="core",
            amount=500.0,
            currency="EUR",
            is_buy=True,
            description="Buy AAPL",
        )

        assert balance.balance == 500.0
        # Buy should use negative amount
        mock_balance_repo.adjust_balance.assert_called_once_with("core", "EUR", -500.0)
        mock_balance_repo.record_transaction.assert_called_once()

    @pytest.mark.asyncio
    async def test_record_trade_settlement_sell(
        self, balance_service, mock_balance_repo, mock_bucket_repo
    ):
        """Test recording a sell trade settlement."""
        mock_bucket_repo.get_by_id.return_value = Bucket(
            id="core",
            name="Core",
            type=BucketType.CORE,
            status=BucketStatus.ACTIVE,
        )
        mock_balance_repo.adjust_balance.return_value = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=1500.0,
            last_updated="2024-01-15",
        )
        mock_balance_repo.record_transaction.return_value = BucketTransaction(
            id=1,
            bucket_id="core",
            type=TransactionType.TRADE_SELL,
            amount=500.0,
            currency="EUR",
        )

        balance = await balance_service.record_trade_settlement(
            bucket_id="core",
            amount=500.0,
            currency="EUR",
            is_buy=False,
            description="Sell AAPL",
        )

        assert balance.balance == 1500.0
        # Sell should use positive amount
        mock_balance_repo.adjust_balance.assert_called_once_with("core", "EUR", 500.0)

    @pytest.mark.asyncio
    async def test_transfer_insufficient_funds(
        self, balance_service, mock_balance_repo, mock_bucket_repo
    ):
        """Test transfer fails with insufficient funds."""
        mock_bucket_repo.get_by_id.return_value = Bucket(
            id="core",
            name="Core",
            type=BucketType.CORE,
            status=BucketStatus.ACTIVE,
        )
        mock_balance_repo.get_balance_amount.return_value = 50.0  # Not enough

        with pytest.raises(ValueError, match="Insufficient funds"):
            await balance_service.transfer_between_buckets(
                from_bucket_id="core",
                to_bucket_id="satellite_1",
                amount=200.0,
                currency="EUR",
            )

    @pytest.mark.asyncio
    async def test_record_dividend(
        self, balance_service, mock_balance_repo, mock_bucket_repo
    ):
        """Test recording a dividend."""
        mock_bucket_repo.get_by_id.return_value = Bucket(
            id="satellite_1",
            name="Satellite",
            type=BucketType.SATELLITE,
            status=BucketStatus.ACTIVE,
        )
        mock_balance_repo.adjust_balance.return_value = BucketBalance(
            bucket_id="satellite_1",
            currency="EUR",
            balance=150.0,
            last_updated="",
        )
        mock_balance_repo.record_transaction.return_value = BucketTransaction(
            id=1,
            bucket_id="satellite_1",
            type=TransactionType.DIVIDEND,
            amount=50.0,
            currency="EUR",
        )

        balance = await balance_service.record_dividend(
            bucket_id="satellite_1",
            amount=50.0,
            currency="EUR",
            description="AAPL dividend",
        )

        assert balance.balance == 150.0
        mock_balance_repo.adjust_balance.assert_called_once_with(
            "satellite_1", "EUR", 50.0
        )
