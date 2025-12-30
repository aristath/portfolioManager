"""Tests for trade sizing service.

These tests validate trade size calculations including lot size constraints,
currency conversion, and minimum trade size enforcement.
"""

import pytest

from app.domain.services.trade_sizing_service import SizedTrade, TradeSizingService


class TestCalculateBuyQuantity:
    """Test calculate_buy_quantity method."""

    def test_calculates_quantity_for_eur_stock(self):
        """Test quantity calculation for EUR stock."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1000.0, price=50.0, min_lot=1, exchange_rate=1.0
        )

        assert result.quantity == 20  # 1000 / 50 = 20 shares
        assert result.value_native == 1000.0
        assert result.value_eur == 1000.0
        assert result.num_lots == 20  # 20 / 1 = 20 lots

    def test_respects_min_lot_constraint(self):
        """Test that minimum lot size is respected."""
        # Target 1000 EUR, price 100 EUR = 10 shares, but min_lot = 20
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1000.0, price=100.0, min_lot=20, exchange_rate=1.0
        )

        assert result.quantity == 20  # Rounded up to min_lot
        assert result.num_lots == 1  # 20 / 20 = 1 lot

    def test_handles_currency_conversion(self):
        """Test handling of currency conversion for non-EUR stocks."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1000.0, price=100.0, min_lot=1, exchange_rate=1.1
        )

        # 1000 EUR / 1.1 = 909.09 USD value
        # 909.09 / 100 = 9.09 shares, rounded down to 9
        assert result.quantity == 9
        assert result.value_native == pytest.approx(900.0, rel=0.01)  # 9 * 100
        assert result.value_eur == pytest.approx(990.0, rel=0.01)  # 900 * 1.1

    def test_rounds_down_to_whole_shares(self):
        """Test that quantity is rounded down to whole shares."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1000.0, price=33.33, min_lot=1, exchange_rate=1.0
        )

        # 1000 / 33.33 = 30.003 shares, should round down to 30
        assert result.quantity == 30
        assert result.value_native == pytest.approx(999.9, rel=0.01)

    def test_handles_fractional_target_that_rounds_to_zero(self):
        """Test handling when target value results in zero shares."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=10.0, price=1000.0, min_lot=1, exchange_rate=1.0
        )

        # 10 / 1000 = 0.01 shares, rounds down to 0
        # But should respect min_lot, so quantity = 1
        assert result.quantity >= 1  # Should be at least min_lot

    def test_calculates_lots_correctly(self):
        """Test that number of lots is calculated correctly."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1000.0, price=50.0, min_lot=10, exchange_rate=1.0
        )

        # 1000 / 50 = 20 shares
        # 20 / 10 = 2 lots
        assert result.quantity == 20
        assert result.num_lots == 2


class TestCalculateSellQuantity:
    """Test calculate_sell_quantity method."""

    def test_calculates_quantity_for_eur_stock(self):
        """Test quantity calculation for EUR stock."""
        result = TradeSizingService.calculate_sell_quantity(
            target_value_eur=500.0, price=50.0, min_lot=1, exchange_rate=1.0
        )

        assert result.quantity == 10  # 500 / 50 = 10 shares
        assert result.value_native == 500.0
        assert result.value_eur == 500.0

    def test_respects_min_lot_constraint(self):
        """Test that minimum lot size is respected for sells."""
        result = TradeSizingService.calculate_sell_quantity(
            target_value_eur=500.0, price=100.0, min_lot=20, exchange_rate=1.0
        )

        # 500 / 100 = 5 shares, but min_lot = 20
        # Should round up to 20
        assert result.quantity == 20

    def test_handles_currency_conversion(self):
        """Test handling of currency conversion for sells."""
        result = TradeSizingService.calculate_sell_quantity(
            target_value_eur=1000.0, price=100.0, min_lot=1, exchange_rate=0.9
        )

        # 1000 EUR / 0.9 = 1111.11 USD value
        # 1111.11 / 100 = 11.11 shares, rounded down to 11
        assert result.quantity == 11
        assert result.value_native == pytest.approx(1100.0, rel=0.01)

    def test_rounds_down_to_whole_shares(self):
        """Test that quantity is rounded down to whole shares for sells."""
        result = TradeSizingService.calculate_sell_quantity(
            target_value_eur=500.0, price=33.33, min_lot=1, exchange_rate=1.0
        )

        # 500 / 33.33 = 15.001 shares, should round down to 15
        assert result.quantity == 15

    def test_calculates_lots_correctly(self):
        """Test that number of lots is calculated correctly for sells."""
        result = TradeSizingService.calculate_sell_quantity(
            target_value_eur=1000.0, price=50.0, min_lot=10, exchange_rate=1.0
        )

        # 1000 / 50 = 20 shares
        # 20 / 10 = 2 lots
        assert result.quantity == 20
        assert result.num_lots == 2


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_handles_zero_price(self):
        """Test handling of zero price (should not crash)."""
        # This would cause division by zero, but method should handle gracefully
        # or raise appropriate error
        with pytest.raises((ZeroDivisionError, ValueError)):
            TradeSizingService.calculate_buy_quantity(
                target_value_eur=1000.0, price=0.0, min_lot=1, exchange_rate=1.0
            )

    def test_handles_zero_target_value(self):
        """Test handling of zero target value."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=0.0, price=50.0, min_lot=1, exchange_rate=1.0
        )

        assert result.quantity == 0
        assert result.value_native == 0.0
        assert result.value_eur == 0.0

    def test_handles_negative_price(self):
        """Test handling of negative price (should not crash)."""
        # Negative price shouldn't happen in practice, but should handle gracefully
        with pytest.raises(ValueError):
            TradeSizingService.calculate_buy_quantity(
                target_value_eur=1000.0, price=-50.0, min_lot=1, exchange_rate=1.0
            )

    def test_handles_very_small_target_value(self):
        """Test handling of very small target values."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1.0, price=100.0, min_lot=1, exchange_rate=1.0
        )

        # 1 / 100 = 0.01 shares, rounds down to 0
        # But min_lot = 1, so should be at least 1
        assert result.quantity >= 1

    def test_handles_very_high_price(self):
        """Test handling of very high stock prices."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1000.0, price=10000.0, min_lot=1, exchange_rate=1.0
        )

        # 1000 / 10000 = 0.1 shares, rounds down to 0
        # But min_lot = 1, so should be 1
        assert result.quantity >= 1

    def test_handles_min_lot_greater_than_calculated_quantity(self):
        """Test when min_lot is greater than calculated quantity."""
        result = TradeSizingService.calculate_buy_quantity(
            target_value_eur=1000.0, price=100.0, min_lot=50, exchange_rate=1.0
        )

        # 1000 / 100 = 10 shares, but min_lot = 50
        # Should round up to 50
        assert result.quantity == 50
        assert result.num_lots == 1

