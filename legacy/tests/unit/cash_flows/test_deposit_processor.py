"""Unit tests for deposit processor."""

import pytest

from app.modules.cash_flows.services.deposit_processor import (
    process_deposit,
    should_process_cash_flow,
)


class TestShouldProcessCashFlow:
    """Test cash flow type detection."""

    @pytest.mark.asyncio
    async def test_detects_deposit_type(self):
        """Test detection of deposit transaction type."""
        cash_flow = {"type": "deposit", "amount": 1000.0}
        assert await should_process_cash_flow(cash_flow) is True

    @pytest.mark.asyncio
    async def test_detects_refill_type(self):
        """Test detection of refill transaction type."""
        cash_flow = {"type": "refill", "amount": 1000.0}
        assert await should_process_cash_flow(cash_flow) is True

    @pytest.mark.asyncio
    async def test_detects_transfer_in_type(self):
        """Test detection of transfer_in transaction type."""
        cash_flow = {"type": "transfer_in", "amount": 1000.0}
        assert await should_process_cash_flow(cash_flow) is True

    @pytest.mark.asyncio
    async def test_ignores_withdrawal_type(self):
        """Test that withdrawals are not processed."""
        cash_flow = {"type": "withdrawal", "amount": 1000.0}
        assert await should_process_cash_flow(cash_flow) is False

    @pytest.mark.asyncio
    async def test_ignores_dividend_type(self):
        """Test that dividends are not processed as deposits."""
        cash_flow = {"type": "dividend", "amount": 100.0}
        assert await should_process_cash_flow(cash_flow) is False

    @pytest.mark.asyncio
    async def test_handles_transaction_type_field(self):
        """Test fallback to transaction_type field."""
        cash_flow = {"transaction_type": "deposit", "amount": 1000.0}
        assert await should_process_cash_flow(cash_flow) is True

    @pytest.mark.asyncio
    async def test_case_insensitive_matching(self):
        """Test case-insensitive transaction type matching."""
        cash_flow = {"type": "DEPOSIT", "amount": 1000.0}
        assert await should_process_cash_flow(cash_flow) is True


class TestProcessDeposit:
    """Test deposit processing logic."""

    @pytest.mark.asyncio
    async def test_processes_deposit_with_satellites(self):
        """Test deposit processing when satellites module is available."""
        result = await process_deposit(
            amount=1000.0,
            currency="EUR",
            transaction_id="TXN123",
            description="Test deposit",
        )

        assert result["total_amount"] == 1000.0
        assert result["currency"] == "EUR"
        assert result["transaction_id"] == "TXN123"
        assert "allocations" in result
        # Should have at least core bucket
        assert "core" in result["allocations"]

    @pytest.mark.asyncio
    async def test_returns_proper_structure(self):
        """Test that result has correct structure."""
        result = await process_deposit(
            amount=1000.0,
            currency="EUR",
            transaction_id="TXN123",
            description="Test deposit",
        )

        # Verify result structure
        assert "total_amount" in result
        assert "currency" in result
        assert "allocations" in result
        assert "transaction_id" in result
        assert isinstance(result["allocations"], dict)
        assert len(result["allocations"]) > 0

    @pytest.mark.asyncio
    async def test_handles_missing_transaction_id(self):
        """Test deposit processing without transaction ID."""
        result = await process_deposit(
            amount=500.0,
            currency="USD",
        )

        assert result["total_amount"] == 500.0
        assert result["currency"] == "USD"
        assert result["transaction_id"] is None
        assert "allocations" in result

    @pytest.mark.asyncio
    async def test_handles_missing_description(self):
        """Test deposit processing without description."""
        result = await process_deposit(
            amount=500.0,
            currency="EUR",
            transaction_id="TXN456",
        )

        assert result["total_amount"] == 500.0
        assert result["transaction_id"] == "TXN456"
        assert "allocations" in result

    @pytest.mark.asyncio
    async def test_processes_usd_deposit(self):
        """Test deposit processing in USD currency."""
        result = await process_deposit(
            amount=1200.0,
            currency="USD",
            transaction_id="TXN789",
            description="USD deposit",
        )

        assert result["total_amount"] == 1200.0
        assert result["currency"] == "USD"
        assert "allocations" in result


class TestDepositProcessorEdgeCases:
    """Test edge cases and error handling."""

    @pytest.mark.asyncio
    async def test_handles_zero_amount(self):
        """Test deposit processing with zero amount."""
        result = await process_deposit(
            amount=0.0,
            currency="EUR",
        )

        assert result["total_amount"] == 0.0
        assert "allocations" in result

    @pytest.mark.asyncio
    async def test_handles_small_amount(self):
        """Test deposit processing with small amount."""
        result = await process_deposit(
            amount=0.01,
            currency="EUR",
        )

        assert result["total_amount"] == 0.01
        assert "allocations" in result

    @pytest.mark.asyncio
    async def test_handles_large_amount(self):
        """Test deposit processing with large amount."""
        result = await process_deposit(
            amount=100000.0,
            currency="EUR",
        )

        assert result["total_amount"] == 100000.0
        assert "allocations" in result
