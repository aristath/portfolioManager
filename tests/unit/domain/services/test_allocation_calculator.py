"""Tests for allocation calculator.

These tests validate allocation calculations for target vs current
allocation comparisons and deviation calculations.
"""

import pytest


class TestCalculateAllocationDeviation:
    """Test calculate_allocation_deviation function."""

    def test_calculates_deviation_correctly(self):
        """Test that deviation is calculated correctly (current - target)."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_deviation,
        )

        target = 0.5  # 50% target
        current = 0.6  # 60% current

        deviation = calculate_allocation_deviation(target, current)

        assert deviation == 0.1  # 60% - 50% = 10% overweight

    def test_calculates_negative_deviation(self):
        """Test that negative deviation is calculated for underweight."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_deviation,
        )

        target = 0.5  # 50% target
        current = 0.4  # 40% current

        deviation = calculate_allocation_deviation(target, current)

        assert deviation == -0.1  # 40% - 50% = -10% underweight

    def test_calculates_zero_deviation_when_equal(self):
        """Test that deviation is zero when current equals target."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_deviation,
        )

        target = 0.5
        current = 0.5

        deviation = calculate_allocation_deviation(target, current)

        assert deviation == 0.0

    def test_handles_zero_target(self):
        """Test handling of zero target."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_deviation,
        )

        target = 0.0
        current = 0.1

        deviation = calculate_allocation_deviation(target, current)

        assert deviation == 0.1  # 10% overweight

    def test_handles_negative_target(self):
        """Test handling of negative target (short position target)."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_deviation,
        )

        target = -0.1  # -10% target (short)
        current = -0.05  # -5% current (less short)

        deviation = calculate_allocation_deviation(target, current)

        assert deviation == 0.05  # Less short = positive deviation

    def test_handles_negative_current(self):
        """Test handling of negative current (short position)."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_deviation,
        )

        target = 0.0
        current = -0.1  # -10% (short position)

        deviation = calculate_allocation_deviation(target, current)

        assert deviation == -0.1  # Short = negative deviation from zero target


class TestCalculateAllocationStatus:
    """Test calculate_allocation_status function."""

    def test_calculates_status_for_overweight(self):
        """Test that overweight status is calculated correctly."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_status,
        )

        target = 0.5
        current = 0.6
        current_value = 6000.0

        status = calculate_allocation_status("country", "US", target, current, current_value)

        assert status.category == "country"
        assert status.name == "US"
        assert status.target_pct == 0.5
        assert status.current_pct == 0.6
        assert status.current_value == 6000.0
        assert status.deviation == 0.1

    def test_calculates_status_for_underweight(self):
        """Test that underweight status is calculated correctly."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_status,
        )

        target = 0.5
        current = 0.4
        current_value = 4000.0

        status = calculate_allocation_status("industry", "Tech", target, current, current_value)

        assert status.category == "industry"
        assert status.name == "Tech"
        assert status.target_pct == 0.5
        assert status.current_pct == 0.4
        assert status.current_value == 4000.0
        assert status.deviation == -0.1

    def test_handles_zero_target(self):
        """Test handling of zero target."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_status,
        )

        target = 0.0
        current = 0.05
        current_value = 500.0

        status = calculate_allocation_status("country", "JP", target, current, current_value)

        assert status.target_pct == 0.0
        assert status.current_pct == 0.05
        assert status.deviation == 0.05

    def test_handles_zero_current(self):
        """Test handling of zero current allocation."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_status,
        )

        target = 0.2
        current = 0.0
        current_value = 0.0

        status = calculate_allocation_status("industry", "Finance", target, current, current_value)

        assert status.target_pct == 0.2
        assert status.current_pct == 0.0
        assert status.deviation == -0.2

    def test_handles_negative_values(self):
        """Test handling of negative target/current (short positions)."""
        from app.domain.services.allocation_calculator import (
            calculate_allocation_status,
        )

        target = -0.1
        current = -0.05
        current_value = -500.0

        status = calculate_allocation_status("country", "UK", target, current, current_value)

        assert status.target_pct == -0.1
        assert status.current_pct == -0.05
        assert status.deviation == 0.05

