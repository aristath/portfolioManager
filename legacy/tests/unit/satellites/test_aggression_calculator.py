"""Unit tests for aggression calculator."""

import pytest

from app.modules.satellites.domain.aggression_calculator import (
    calculate_aggression,
    get_aggression_description,
    scale_position_size,
    should_hibernate,
)


class TestCalculateAggression:
    """Test aggression calculation logic."""

    def test_full_aggression_at_target(self):
        """Test full aggression when at 100% of target with no drawdown."""
        result = calculate_aggression(
            current_value=10000.0,
            target_value=10000.0,
            high_water_mark=10000.0,
        )

        assert result.aggression == pytest.approx(1.0)
        assert result.allocation_aggression == pytest.approx(1.0)
        assert result.drawdown_aggression == pytest.approx(1.0)
        assert result.pct_of_target == pytest.approx(1.0)
        assert result.drawdown == pytest.approx(0.0)
        assert result.in_hibernation is False

    def test_full_aggression_above_target(self):
        """Test full aggression when above target."""
        result = calculate_aggression(
            current_value=15000.0,
            target_value=10000.0,
            high_water_mark=15000.0,
        )

        assert result.aggression == pytest.approx(1.0)
        assert result.allocation_aggression == pytest.approx(1.0)
        assert result.pct_of_target == pytest.approx(1.5)

    def test_reduced_aggression_at_90_percent(self):
        """Test 80% aggression when at 90% of target."""
        result = calculate_aggression(
            current_value=9000.0,
            target_value=10000.0,
            high_water_mark=9000.0,
        )

        assert result.aggression == pytest.approx(0.8)
        assert result.allocation_aggression == pytest.approx(0.8)
        assert result.pct_of_target == pytest.approx(0.9)

    def test_conservative_at_70_percent(self):
        """Test 60% aggression when at 70% of target."""
        result = calculate_aggression(
            current_value=7000.0,
            target_value=10000.0,
            high_water_mark=7000.0,
        )

        assert result.aggression == pytest.approx(0.6)
        assert result.allocation_aggression == pytest.approx(0.6)
        assert result.pct_of_target == pytest.approx(0.7)

    def test_very_conservative_at_50_percent(self):
        """Test 40% aggression when at 50% of target."""
        result = calculate_aggression(
            current_value=5000.0,
            target_value=10000.0,
            high_water_mark=5000.0,
        )

        assert result.aggression == pytest.approx(0.4)
        assert result.allocation_aggression == pytest.approx(0.4)
        assert result.pct_of_target == pytest.approx(0.5)

    def test_hibernation_below_40_percent(self):
        """Test hibernation when below 40% of target."""
        result = calculate_aggression(
            current_value=3000.0,
            target_value=10000.0,
            high_water_mark=3000.0,
        )

        assert result.aggression == pytest.approx(0.0)
        assert result.allocation_aggression == pytest.approx(0.0)
        assert result.pct_of_target == pytest.approx(0.3)
        assert result.in_hibernation is True
        assert result.limiting_factor == "allocation"

    def test_drawdown_10_percent_full_aggression(self):
        """Test full aggression with 10% drawdown (below threshold)."""
        result = calculate_aggression(
            current_value=9000.0,
            target_value=8000.0,  # Above target but in drawdown
            high_water_mark=10000.0,
        )

        # Allocation: 9000/8000 = 1.125 → 1.0
        # Drawdown: (10000-9000)/10000 = 0.1 → 1.0
        assert result.aggression == pytest.approx(1.0)
        assert result.allocation_aggression == pytest.approx(1.0)
        assert result.drawdown_aggression == pytest.approx(1.0)
        assert result.drawdown == pytest.approx(0.1)

    def test_drawdown_20_percent_reduced(self):
        """Test reduced aggression (70%) with 20% drawdown."""
        result = calculate_aggression(
            current_value=8000.0,
            target_value=8000.0,
            high_water_mark=10000.0,
        )

        # Allocation: 8000/8000 = 1.0 → 1.0
        # Drawdown: (10000-8000)/10000 = 0.2 → 0.7
        # Min(1.0, 0.7) = 0.7
        assert result.aggression == pytest.approx(0.7)
        assert result.allocation_aggression == pytest.approx(1.0)
        assert result.drawdown_aggression == pytest.approx(0.7)
        assert result.drawdown == pytest.approx(0.2)
        assert result.limiting_factor == "drawdown"

    def test_drawdown_30_percent_minimal(self):
        """Test minimal aggression (30%) with 30% drawdown."""
        result = calculate_aggression(
            current_value=7000.0,
            target_value=7000.0,
            high_water_mark=10000.0,
        )

        # Drawdown: (10000-7000)/10000 = 0.3 → 0.3
        assert result.aggression == pytest.approx(0.3)
        assert result.drawdown_aggression == pytest.approx(0.3)
        assert result.drawdown == pytest.approx(0.3)

    def test_drawdown_40_percent_hibernation(self):
        """Test hibernation with 40% drawdown (severe)."""
        result = calculate_aggression(
            current_value=6000.0,
            target_value=10000.0,
            high_water_mark=10000.0,
        )

        # Drawdown: (10000-6000)/10000 = 0.4 → 0.0
        assert result.aggression == pytest.approx(0.0)
        assert result.drawdown_aggression == pytest.approx(0.0)
        assert result.drawdown == pytest.approx(0.4)
        assert result.in_hibernation is True
        assert result.limiting_factor == "drawdown"

    def test_most_conservative_wins(self):
        """Test that most conservative factor wins."""
        result = calculate_aggression(
            current_value=9000.0,  # 90% of target → 0.8 aggression
            target_value=10000.0,
            high_water_mark=12000.0,  # 25% drawdown → 0.3 aggression
        )

        # Allocation: 9000/10000 = 0.9 → 0.8
        # Drawdown: (12000-9000)/12000 = 0.25 → 0.3
        # Min(0.8, 0.3) = 0.3
        assert result.aggression == pytest.approx(0.3)
        assert result.allocation_aggression == pytest.approx(0.8)
        assert result.drawdown_aggression == pytest.approx(0.3)
        assert result.limiting_factor == "drawdown"

    def test_no_high_water_mark(self):
        """Test calculation when no high water mark provided."""
        result = calculate_aggression(
            current_value=8000.0,
            target_value=10000.0,
            high_water_mark=None,
        )

        # No drawdown → full drawdown aggression
        assert result.drawdown == pytest.approx(0.0)
        assert result.drawdown_aggression == pytest.approx(1.0)
        # Limited by allocation only
        assert result.aggression == pytest.approx(0.8)  # 80% of target

    def test_zero_target_value(self):
        """Test handling of zero target value."""
        result = calculate_aggression(
            current_value=5000.0,
            target_value=0.0,
            high_water_mark=5000.0,
        )

        # Zero target → pct_of_target = 0 → hibernation
        assert result.pct_of_target == pytest.approx(0.0)
        assert result.allocation_aggression == pytest.approx(0.0)
        assert result.in_hibernation is True


class TestShouldHibernate:
    """Test hibernation detection."""

    def test_should_hibernate_when_aggression_zero(self):
        """Test hibernation detection when aggression is 0.0."""
        result = calculate_aggression(
            current_value=3000.0,
            target_value=10000.0,
            high_water_mark=10000.0,
        )

        assert should_hibernate(result) is True

    def test_should_not_hibernate_when_aggressive(self):
        """Test no hibernation when aggression > 0."""
        result = calculate_aggression(
            current_value=10000.0,
            target_value=10000.0,
            high_water_mark=10000.0,
        )

        assert should_hibernate(result) is False


class TestScalePositionSize:
    """Test position size scaling."""

    def test_scale_full_aggression(self):
        """Test scaling with full aggression (100%)."""
        assert scale_position_size(1000.0, 1.0) == pytest.approx(1000.0)

    def test_scale_80_percent(self):
        """Test scaling with 80% aggression."""
        assert scale_position_size(1000.0, 0.8) == pytest.approx(800.0)

    def test_scale_50_percent(self):
        """Test scaling with 50% aggression."""
        assert scale_position_size(1000.0, 0.5) == pytest.approx(500.0)

    def test_scale_hibernation(self):
        """Test scaling with 0% aggression (hibernation)."""
        assert scale_position_size(1000.0, 0.0) == pytest.approx(0.0)


class TestGetAggressionDescription:
    """Test aggression description generation."""

    def test_description_full_aggression(self):
        """Test description for full aggression."""
        result = calculate_aggression(
            current_value=10000.0,
            target_value=10000.0,
            high_water_mark=10000.0,
        )

        desc = get_aggression_description(result)

        assert "Fully funded (100.0% of target)" in desc
        assert "No drawdown" in desc
        assert "Aggression: 100%" in desc

    def test_description_partial_funding(self):
        """Test description for partial funding."""
        result = calculate_aggression(
            current_value=7000.0,
            target_value=10000.0,
            high_water_mark=7000.0,
        )

        desc = get_aggression_description(result)

        assert "Funding: 70.0% of target" in desc
        assert "Aggression: 60%" in desc

    def test_description_with_drawdown(self):
        """Test description with drawdown."""
        result = calculate_aggression(
            current_value=8000.0,
            target_value=8000.0,
            high_water_mark=10000.0,
        )

        desc = get_aggression_description(result)

        assert "Drawdown: 20.0%" in desc
        assert "limited by drawdown" in desc

    def test_description_hibernation_allocation(self):
        """Test description for hibernation due to allocation."""
        result = calculate_aggression(
            current_value=3000.0,
            target_value=10000.0,
            high_water_mark=3000.0,
        )

        desc = get_aggression_description(result)

        assert "HIBERNATION" in desc
        assert "30.0% of target" in desc
        assert "below 40% threshold" in desc

    def test_description_hibernation_drawdown(self):
        """Test description for hibernation due to drawdown."""
        result = calculate_aggression(
            current_value=6000.0,
            target_value=10000.0,
            high_water_mark=10000.0,
        )

        desc = get_aggression_description(result)

        assert "HIBERNATION" in desc
        assert "Drawdown at 40.0%" in desc
        assert "above 35% threshold" in desc
