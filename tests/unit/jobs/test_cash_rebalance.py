"""Tests for cash rebalance job.

These tests validate the drip execution strategy that executes
one trade per cycle with proper P&L guardrails.
CRITICAL: This is the core trading execution job.
"""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest


class TestCheckPnlGuardrails:
    """Test P&L guardrail checks."""

    @pytest.mark.asyncio
    async def test_allows_trading_when_pnl_ok(self):
        """Test that trading is allowed when P&L status is OK."""
        from app.jobs.cash_rebalance import _check_pnl_guardrails

        mock_tracker = AsyncMock()
        mock_tracker.get_trading_status.return_value = {
            "status": "ok",
            "pnl_display": "+€50.00",
            "can_buy": True,
            "can_sell": True,
            "reason": None,
        }

        with patch(
            "app.jobs.cash_rebalance.get_daily_pnl_tracker",
            return_value=mock_tracker,
        ):
            pnl_status, can_trade = await _check_pnl_guardrails()

        assert can_trade is True
        assert pnl_status["status"] == "ok"

    @pytest.mark.asyncio
    async def test_blocks_trading_when_halted(self):
        """Test that trading is blocked when P&L status is halted."""
        from app.jobs.cash_rebalance import _check_pnl_guardrails

        mock_tracker = AsyncMock()
        mock_tracker.get_trading_status.return_value = {
            "status": "halted",
            "pnl_display": "-€500.00",
            "can_buy": False,
            "can_sell": False,
            "reason": "Daily loss limit exceeded",
        }

        with patch(
            "app.jobs.cash_rebalance.get_daily_pnl_tracker",
            return_value=mock_tracker,
        ):
            with patch("app.jobs.cash_rebalance.emit"):
                with patch("app.jobs.cash_rebalance.set_error"):
                    pnl_status, can_trade = await _check_pnl_guardrails()

        assert can_trade is False
        assert pnl_status["status"] == "halted"


class TestValidateNextAction:
    """Test next action validation."""

    @pytest.fixture
    def mock_trade_repo(self):
        """Create mock trade repository."""
        return AsyncMock()

    @pytest.fixture
    def mock_buy_action(self):
        """Create mock BUY action."""
        from app.domain.value_objects.trade_side import TradeSide

        action = MagicMock()
        action.side = TradeSide.BUY
        action.symbol = "AAPL"
        return action

    @pytest.fixture
    def mock_sell_action(self):
        """Create mock SELL action."""
        from app.domain.value_objects.trade_side import TradeSide

        action = MagicMock()
        action.side = TradeSide.SELL
        action.symbol = "AAPL"
        return action

    @pytest.mark.asyncio
    async def test_allows_buy_when_cash_sufficient(
        self, mock_trade_repo, mock_buy_action
    ):
        """Test BUY is allowed when cash is sufficient."""
        from app.jobs.cash_rebalance import _validate_next_action

        pnl_status = {"can_buy": True, "can_sell": True}
        cash_balance = 1000.0
        min_trade_size = 500.0

        result = await _validate_next_action(
            mock_buy_action,
            pnl_status,
            cash_balance,
            min_trade_size,
            mock_trade_repo,
        )

        assert result is True

    @pytest.mark.asyncio
    async def test_blocks_buy_when_cash_insufficient(
        self, mock_trade_repo, mock_buy_action
    ):
        """Test BUY is blocked when cash is insufficient."""
        from app.jobs.cash_rebalance import _validate_next_action

        pnl_status = {"can_buy": True, "can_sell": True}
        cash_balance = 400.0
        min_trade_size = 500.0

        result = await _validate_next_action(
            mock_buy_action,
            pnl_status,
            cash_balance,
            min_trade_size,
            mock_trade_repo,
        )

        assert result is False

    @pytest.mark.asyncio
    async def test_blocks_buy_when_pnl_guardrail_active(
        self, mock_trade_repo, mock_buy_action
    ):
        """Test BUY is blocked by P&L guardrail."""
        from app.jobs.cash_rebalance import _validate_next_action

        pnl_status = {"can_buy": False, "can_sell": True, "reason": "Daily loss limit"}
        cash_balance = 1000.0
        min_trade_size = 500.0

        result = await _validate_next_action(
            mock_buy_action,
            pnl_status,
            cash_balance,
            min_trade_size,
            mock_trade_repo,
        )

        assert result is False

    @pytest.mark.asyncio
    async def test_blocks_sell_when_pnl_guardrail_active(
        self, mock_trade_repo, mock_sell_action
    ):
        """Test SELL is blocked by P&L guardrail."""
        from app.jobs.cash_rebalance import _validate_next_action

        pnl_status = {"can_buy": True, "can_sell": False, "reason": "Max loss reached"}
        cash_balance = 1000.0
        min_trade_size = 500.0

        result = await _validate_next_action(
            mock_sell_action,
            pnl_status,
            cash_balance,
            min_trade_size,
            mock_trade_repo,
        )

        assert result is False

    @pytest.mark.asyncio
    async def test_blocks_sell_when_recent_order_exists(
        self, mock_trade_repo, mock_sell_action
    ):
        """Test SELL is blocked when recent sell order exists."""
        from app.jobs.cash_rebalance import _validate_next_action

        pnl_status = {"can_buy": True, "can_sell": True}
        cash_balance = 1000.0
        min_trade_size = 500.0
        mock_trade_repo.has_recent_sell_order.return_value = True

        result = await _validate_next_action(
            mock_sell_action,
            pnl_status,
            cash_balance,
            min_trade_size,
            mock_trade_repo,
        )

        assert result is False

    @pytest.mark.asyncio
    async def test_allows_sell_when_no_recent_order(
        self, mock_trade_repo, mock_sell_action
    ):
        """Test SELL is allowed when no recent order exists."""
        from app.jobs.cash_rebalance import _validate_next_action

        pnl_status = {"can_buy": True, "can_sell": True}
        cash_balance = 1000.0
        min_trade_size = 500.0
        mock_trade_repo.has_recent_sell_order.return_value = False

        result = await _validate_next_action(
            mock_sell_action,
            pnl_status,
            cash_balance,
            min_trade_size,
            mock_trade_repo,
        )

        assert result is True
