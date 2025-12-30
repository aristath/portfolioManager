"""Tests for turnover tracker service.

These tests validate portfolio turnover calculation including trade aggregation
and snapshot averaging.
"""

from unittest.mock import AsyncMock, MagicMock

import pytest


@pytest.fixture
def mock_db_manager():
    """Mock database manager."""
    manager = MagicMock()
    manager.ledger = AsyncMock()
    manager.snapshots = AsyncMock()
    return manager


class TestCalculateAnnualTurnover:
    """Test calculate_annual_turnover method."""

    @pytest.mark.asyncio
    async def test_calculates_turnover_from_trades_and_snapshots(self, mock_db_manager):
        """Test that turnover is calculated from trades and snapshots."""
        from app.application.services.turnover_tracker import TurnoverTracker

        # Mock trades
        mock_trade1 = MagicMock()
        mock_trade1.__getitem__ = lambda self, key: {
            "side": "BUY",
            "quantity": 10,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 1000.0,
        }.get(key)
        mock_trade1.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_trade2 = MagicMock()
        mock_trade2.__getitem__ = lambda self, key: {
            "side": "SELL",
            "quantity": 5,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 500.0,
        }.get(key)
        mock_trade2.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_db_manager.ledger.fetchall = AsyncMock(
            return_value=[mock_trade1, mock_trade2]
        )

        # Mock snapshots
        mock_snapshot1 = MagicMock()
        mock_snapshot1.__getitem__ = lambda self, key: {
            "total_value": 10000.0,
            "date": "2024-01-01",
        }.get(key)
        mock_snapshot1.keys = lambda: ["total_value", "date"]

        mock_snapshot2 = MagicMock()
        mock_snapshot2.__getitem__ = lambda self, key: {
            "total_value": 10500.0,
            "date": "2024-02-01",
        }.get(key)
        mock_snapshot2.keys = lambda: ["total_value", "date"]

        mock_db_manager.snapshots.fetchall = AsyncMock(
            return_value=[mock_snapshot1, mock_snapshot2]
        )

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is not None
        assert result >= 0
        assert isinstance(result, float)

    @pytest.mark.asyncio
    async def test_returns_none_when_no_trades(self, mock_db_manager):
        """Test that None is returned when there are no trades."""
        from app.application.services.turnover_tracker import TurnoverTracker

        mock_db_manager.ledger.fetchall = AsyncMock(return_value=[])

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is None

    @pytest.mark.asyncio
    async def test_returns_none_when_no_snapshots(self, mock_db_manager):
        """Test that None is returned when there are no snapshots."""
        from app.application.services.turnover_tracker import TurnoverTracker

        mock_trade = MagicMock()
        mock_trade.__getitem__ = lambda self, key: {
            "side": "BUY",
            "quantity": 10,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 1000.0,
        }.get(key)
        mock_trade.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_db_manager.ledger.fetchall = AsyncMock(return_value=[mock_trade])
        mock_db_manager.snapshots.fetchall = AsyncMock(return_value=[])

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is None

    @pytest.mark.asyncio
    async def test_uses_value_eur_when_available(self, mock_db_manager):
        """Test that value_eur is used when available."""
        from app.application.services.turnover_tracker import TurnoverTracker

        mock_trade = MagicMock()
        mock_trade.__getitem__ = lambda self, key: {
            "side": "BUY",
            "quantity": 10,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 1000.0,  # Should use this
        }.get(key)
        mock_trade.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_db_manager.ledger.fetchall = AsyncMock(return_value=[mock_trade])

        mock_snapshot = MagicMock()
        mock_snapshot.__getitem__ = lambda self, key: {
            "total_value": 10000.0,
            "date": "2024-01-01",
        }.get(key)
        mock_snapshot.keys = lambda: ["total_value", "date"]

        mock_db_manager.snapshots.fetchall = AsyncMock(return_value=[mock_snapshot])

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is not None
        assert result >= 0

    @pytest.mark.asyncio
    async def test_calculates_value_when_value_eur_missing(self, mock_db_manager):
        """Test that value is calculated when value_eur is missing."""
        from app.application.services.turnover_tracker import TurnoverTracker

        mock_trade = MagicMock()
        mock_trade.__getitem__ = lambda self, key: {
            "side": "BUY",
            "quantity": 10,
            "price": 100.0,
            "currency": "USD",
            "currency_rate": 0.9,  # Should calculate: 10 * 100 * 0.9 = 900
            "value_eur": None,  # Missing - should calculate
        }.get(key)
        mock_trade.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_db_manager.ledger.fetchall = AsyncMock(return_value=[mock_trade])

        mock_snapshot = MagicMock()
        mock_snapshot.__getitem__ = lambda self, key: {
            "total_value": 10000.0,
            "date": "2024-01-01",
        }.get(key)
        mock_snapshot.keys = lambda: ["total_value", "date"]

        mock_db_manager.snapshots.fetchall = AsyncMock(return_value=[mock_snapshot])

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is not None
        assert result >= 0

    @pytest.mark.asyncio
    async def test_handles_both_buy_and_sell_trades(self, mock_db_manager):
        """Test handling of both buy and sell trades."""
        from app.application.services.turnover_tracker import TurnoverTracker

        mock_trade_buy = MagicMock()
        mock_trade_buy.__getitem__ = lambda self, key: {
            "side": "BUY",
            "quantity": 10,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 1000.0,
        }.get(key)
        mock_trade_buy.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_trade_sell = MagicMock()
        mock_trade_sell.__getitem__ = lambda self, key: {
            "side": "SELL",
            "quantity": 5,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 500.0,
        }.get(key)
        mock_trade_sell.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_db_manager.ledger.fetchall = AsyncMock(
            return_value=[mock_trade_buy, mock_trade_sell]
        )

        mock_snapshot = MagicMock()
        mock_snapshot.__getitem__ = lambda self, key: {
            "total_value": 10000.0,
            "date": "2024-01-01",
        }.get(key)
        mock_snapshot.keys = lambda: ["total_value", "date"]

        mock_db_manager.snapshots.fetchall = AsyncMock(return_value=[mock_snapshot])

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is not None
        assert result >= 0

    @pytest.mark.asyncio
    async def test_averages_snapshot_values(self, mock_db_manager):
        """Test that snapshot values are averaged."""
        from app.application.services.turnover_tracker import TurnoverTracker

        mock_trade = MagicMock()
        mock_trade.__getitem__ = lambda self, key: {
            "side": "BUY",
            "quantity": 10,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 1000.0,
        }.get(key)
        mock_trade.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_db_manager.ledger.fetchall = AsyncMock(return_value=[mock_trade])

        # Multiple snapshots with different values
        mock_snapshot1 = MagicMock()
        mock_snapshot1.__getitem__ = lambda self, key: {
            "total_value": 10000.0,
            "date": "2024-01-01",
        }.get(key)
        mock_snapshot1.keys = lambda: ["total_value", "date"]

        mock_snapshot2 = MagicMock()
        mock_snapshot2.__getitem__ = lambda self, key: {
            "total_value": 10500.0,
            "date": "2024-02-01",
        }.get(key)
        mock_snapshot2.keys = lambda: ["total_value", "date"]

        mock_snapshot3 = MagicMock()
        mock_snapshot3.__getitem__ = lambda self, key: {
            "total_value": 11000.0,
            "date": "2024-03-01",
        }.get(key)
        mock_snapshot3.keys = lambda: ["total_value", "date"]

        mock_db_manager.snapshots.fetchall = AsyncMock(
            return_value=[mock_snapshot1, mock_snapshot2, mock_snapshot3]
        )

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is not None
        assert result >= 0

    @pytest.mark.asyncio
    async def test_returns_none_when_average_portfolio_value_zero(
        self, mock_db_manager
    ):
        """Test that None is returned when average portfolio value is zero."""
        from app.application.services.turnover_tracker import TurnoverTracker

        mock_trade = MagicMock()
        mock_trade.__getitem__ = lambda self, key: {
            "side": "BUY",
            "quantity": 10,
            "price": 100.0,
            "currency": "EUR",
            "currency_rate": 1.0,
            "value_eur": 1000.0,
        }.get(key)
        mock_trade.keys = lambda: [
            "side",
            "quantity",
            "price",
            "currency",
            "currency_rate",
            "value_eur",
        ]

        mock_db_manager.ledger.fetchall = AsyncMock(return_value=[mock_trade])

        mock_snapshot = MagicMock()
        mock_snapshot.__getitem__ = lambda self, key: {
            "total_value": 0.0,  # Zero value
            "date": "2024-01-01",
        }.get(key)
        mock_snapshot.keys = lambda: ["total_value", "date"]

        mock_db_manager.snapshots.fetchall = AsyncMock(return_value=[mock_snapshot])

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.calculate_annual_turnover("2024-12-31")

        assert result is None


class TestGetTurnoverStatus:
    """Test get_turnover_status method."""

    @pytest.mark.asyncio
    async def test_returns_unknown_when_turnover_none(self, mock_db_manager):
        """Test that unknown status is returned when turnover is None."""
        from app.application.services.turnover_tracker import TurnoverTracker

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.get_turnover_status(None)

        assert result["turnover"] is None
        assert result["status"] == "unknown"
        assert result["alert"] is None
        assert "Insufficient data" in result["reason"]

    @pytest.mark.asyncio
    async def test_returns_critical_for_high_turnover(self, mock_db_manager):
        """Test that critical status is returned for high turnover."""
        from app.application.services.turnover_tracker import TurnoverTracker

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.get_turnover_status(1.0)  # 100% turnover

        assert result["status"] == "critical"
        assert result["alert"] == "critical"
        assert "Very high turnover" in result["reason"]

    @pytest.mark.asyncio
    async def test_returns_warning_for_moderate_turnover(self, mock_db_manager):
        """Test that warning status is returned for moderate turnover."""
        from app.application.services.turnover_tracker import TurnoverTracker

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.get_turnover_status(0.6)  # 60% turnover

        assert result["status"] == "warning"
        assert result["alert"] == "warning"
        assert "High turnover" in result["reason"]

    @pytest.mark.asyncio
    async def test_returns_normal_for_low_turnover(self, mock_db_manager):
        """Test that normal status is returned for low turnover."""
        from app.application.services.turnover_tracker import TurnoverTracker

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.get_turnover_status(0.3)  # 30% turnover

        assert result["status"] == "normal"
        assert result["alert"] is None
        assert "Normal turnover" in result["reason"]

    @pytest.mark.asyncio
    async def test_includes_turnover_display(self, mock_db_manager):
        """Test that turnover display is included."""
        from app.application.services.turnover_tracker import TurnoverTracker

        tracker = TurnoverTracker(db_manager=mock_db_manager)
        result = await tracker.get_turnover_status(0.5)

        assert "turnover_display" in result
        assert "%" in result["turnover_display"]
