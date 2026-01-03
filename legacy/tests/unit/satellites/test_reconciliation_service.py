"""Unit tests for ReconciliationService."""

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
from app.modules.satellites.services.reconciliation_service import ReconciliationService


@pytest.fixture
def mock_balance_repo():
    """Create a mock balance repository."""
    repo = MagicMock()
    repo.get_total_by_currency = AsyncMock()
    repo.get_balance_amount = AsyncMock()
    repo.set_balance = AsyncMock()
    repo.adjust_balance = AsyncMock()
    repo.record_transaction = AsyncMock()
    repo.get_recent_transactions = AsyncMock(return_value=[])
    return repo


@pytest.fixture
def mock_bucket_repo():
    """Create a mock bucket repository."""
    repo = MagicMock()
    repo.get_all = AsyncMock()
    return repo


@pytest.fixture
def reconciliation_service(mock_balance_repo, mock_bucket_repo):
    """Create a ReconciliationService with mocked dependencies."""
    return ReconciliationService(
        balance_repo=mock_balance_repo,
        bucket_repo=mock_bucket_repo,
    )


class TestReconciliationService:
    """Tests for ReconciliationService."""

    @pytest.mark.asyncio
    async def test_check_invariant_reconciled(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test check_invariant when balances match."""
        mock_balance_repo.get_total_by_currency.return_value = 1000.0

        result = await reconciliation_service.check_invariant(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert result.is_reconciled is True
        assert result.difference == 0.0
        assert result.virtual_total == 1000.0
        assert result.actual_total == 1000.0

    @pytest.mark.asyncio
    async def test_check_invariant_not_reconciled(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test check_invariant when balances don't match."""
        mock_balance_repo.get_total_by_currency.return_value = 1050.0

        result = await reconciliation_service.check_invariant(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert result.is_reconciled is False
        assert result.difference == 50.0  # Virtual is 50 more than actual
        assert result.virtual_total == 1050.0
        assert result.actual_total == 1000.0

    @pytest.mark.asyncio
    async def test_check_invariant_allows_small_difference(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test check_invariant allows sub-cent differences."""
        mock_balance_repo.get_total_by_currency.return_value = 1000.005

        result = await reconciliation_service.check_invariant(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert result.is_reconciled is True  # 0.005 < 0.01 threshold

    @pytest.mark.asyncio
    async def test_reconcile_already_reconciled(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test reconcile returns immediately if already reconciled."""
        mock_balance_repo.get_total_by_currency.return_value = 1000.0

        result = await reconciliation_service.reconcile(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert result.is_reconciled is True
        assert result.adjustments_made == {}
        mock_balance_repo.adjust_balance.assert_not_called()

    @pytest.mark.asyncio
    async def test_reconcile_auto_correct_small_difference(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test reconcile auto-corrects small differences."""
        mock_balance_repo.get_total_by_currency.return_value = 1000.50  # 0.50 over
        mock_balance_repo.adjust_balance.return_value = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=999.50,
            last_updated="",
        )
        mock_balance_repo.record_transaction.return_value = BucketTransaction(
            id=1,
            bucket_id="core",
            type=TransactionType.REALLOCATION,
            amount=-0.50,
            currency="EUR",
        )

        result = await reconciliation_service.reconcile(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert result.is_reconciled is True
        assert "core" in result.adjustments_made
        assert result.adjustments_made["core"] == -0.50
        mock_balance_repo.adjust_balance.assert_called_once()

    @pytest.mark.asyncio
    async def test_reconcile_large_difference_not_auto_corrected(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test reconcile does not auto-correct large differences."""
        mock_balance_repo.get_total_by_currency.return_value = 1100.0  # 100 over

        result = await reconciliation_service.reconcile(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert result.is_reconciled is False
        assert result.adjustments_made == {}
        mock_balance_repo.adjust_balance.assert_not_called()

    @pytest.mark.asyncio
    async def test_reconcile_custom_threshold(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test reconcile with custom auto-correct threshold."""
        mock_balance_repo.get_total_by_currency.return_value = 1005.0  # 5 over
        mock_balance_repo.adjust_balance.return_value = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=995.0,
            last_updated="",
        )
        mock_balance_repo.record_transaction.return_value = BucketTransaction(
            id=1,
            bucket_id="core",
            type=TransactionType.REALLOCATION,
            amount=-5.0,
            currency="EUR",
        )

        result = await reconciliation_service.reconcile(
            currency="EUR",
            actual_balance=1000.0,
            auto_correct_threshold=10.0,  # Higher threshold
        )

        assert result.is_reconciled is True
        assert result.adjustments_made["core"] == -5.0


class TestReconciliationServiceBulk:
    """Tests for bulk reconciliation operations."""

    @pytest.mark.asyncio
    async def test_reconcile_all(self, reconciliation_service, mock_balance_repo):
        """Test reconciling all currencies."""
        mock_balance_repo.get_total_by_currency.side_effect = [1000.0, 500.0]

        actual_balances = {"EUR": 1000.0, "USD": 500.0}

        results = await reconciliation_service.reconcile_all(actual_balances)

        assert len(results) == 2
        assert all(r.is_reconciled for r in results)

    @pytest.mark.asyncio
    async def test_initialize_from_brokerage(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test initializing balances from brokerage."""
        # First call returns 0 (no virtual balances)
        # Second call after set_balance returns actual balance
        mock_balance_repo.get_total_by_currency.side_effect = [0.0, 1000.0]
        mock_balance_repo.set_balance.return_value = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=1000.0,
            last_updated="",
        )
        mock_balance_repo.record_transaction.return_value = BucketTransaction(
            id=1,
            bucket_id="core",
            type=TransactionType.DEPOSIT,
            amount=1000.0,
            currency="EUR",
        )

        results = await reconciliation_service.initialize_from_brokerage(
            {"EUR": 1000.0}
        )

        assert len(results) == 1
        mock_balance_repo.set_balance.assert_called_once_with("core", "EUR", 1000.0)
        mock_balance_repo.record_transaction.assert_called_once()


class TestForceReconciliation:
    """Tests for force reconciliation."""

    @pytest.mark.asyncio
    async def test_force_reconcile_to_core(
        self, reconciliation_service, mock_balance_repo, mock_bucket_repo
    ):
        """Test force reconciliation adjusts core balance."""
        # Setup: Core has 800, satellite has 300 = 1100 virtual
        # Actual is 1000, so core should become 700
        core = Bucket(
            id="core",
            name="Core",
            type=BucketType.CORE,
            status=BucketStatus.ACTIVE,
        )
        satellite = Bucket(
            id="satellite_1",
            name="Satellite",
            type=BucketType.SATELLITE,
            status=BucketStatus.ACTIVE,
        )
        mock_bucket_repo.get_all.return_value = [core, satellite]

        mock_balance_repo.get_balance_amount.side_effect = [
            300.0,  # satellite balance
            800.0,  # core balance
        ]
        mock_balance_repo.set_balance.return_value = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=700.0,
            last_updated="",
        )
        mock_balance_repo.record_transaction.return_value = BucketTransaction(
            id=1,
            bucket_id="core",
            type=TransactionType.REALLOCATION,
            amount=-100.0,
            currency="EUR",
        )

        result = await reconciliation_service.force_reconcile_to_core(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert result.is_reconciled is True
        assert result.difference == 0.0
        mock_balance_repo.set_balance.assert_called_once_with("core", "EUR", 700.0)


class TestDiagnoseDiscrepancy:
    """Tests for discrepancy diagnosis."""

    @pytest.mark.asyncio
    async def test_diagnose_discrepancy(
        self, reconciliation_service, mock_balance_repo, mock_bucket_repo
    ):
        """Test diagnosing a balance discrepancy."""
        core = Bucket(
            id="core",
            name="Core",
            type=BucketType.CORE,
            status=BucketStatus.ACTIVE,
        )
        mock_bucket_repo.get_all.return_value = [core]
        mock_balance_repo.get_balance_amount.return_value = 800.0
        mock_balance_repo.get_recent_transactions.return_value = [
            BucketTransaction(
                id=1,
                bucket_id="core",
                type=TransactionType.DEPOSIT,
                amount=1000.0,
                currency="EUR",
                created_at="2024-01-15",
            ),
            BucketTransaction(
                id=2,
                bucket_id="core",
                type=TransactionType.TRADE_BUY,
                amount=-200.0,
                currency="EUR",
                created_at="2024-01-16",
            ),
        ]

        diagnosis = await reconciliation_service.diagnose_discrepancy(
            currency="EUR",
            actual_balance=1000.0,
        )

        assert "currency" in diagnosis
        assert diagnosis["currency"] == "EUR"
        assert "breakdown" in diagnosis
        assert "recent_transactions" in diagnosis
        assert diagnosis["virtual_total"] == 800.0
        assert diagnosis["actual_balance"] == 1000.0
        assert diagnosis["difference"] == -200.0


class TestReconciliationResult:
    """Tests for ReconciliationResult."""

    @pytest.mark.asyncio
    async def test_difference_pct_calculation(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test difference_pct property calculation."""
        mock_balance_repo.get_total_by_currency.return_value = 1100.0

        result = await reconciliation_service.check_invariant(
            currency="EUR",
            actual_balance=1000.0,
        )

        # Difference is 100, actual is 1000, so 10%
        assert result.difference_pct == 0.1

    @pytest.mark.asyncio
    async def test_difference_pct_zero_actual(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test difference_pct when actual is zero."""
        mock_balance_repo.get_total_by_currency.return_value = 100.0

        result = await reconciliation_service.check_invariant(
            currency="EUR",
            actual_balance=0.0,
        )

        # Division by zero should return infinity
        assert result.difference_pct == float("inf")

    @pytest.mark.asyncio
    async def test_difference_pct_both_zero(
        self, reconciliation_service, mock_balance_repo
    ):
        """Test difference_pct when both are zero."""
        mock_balance_repo.get_total_by_currency.return_value = 0.0

        result = await reconciliation_service.check_invariant(
            currency="EUR",
            actual_balance=0.0,
        )

        assert result.difference_pct == 0.0
