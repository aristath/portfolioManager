"""Unit tests for priority calculator."""

import pytest

from app.domain.services.priority_calculator import (
    PriorityCalculator,
    PriorityInput,
)
from app.domain.utils.priority_helpers import (
    calculate_weight_boost,
    calculate_risk_adjustment,
)


class TestPriorityHelpers:
    """Tests for priority helper functions."""

    def test_calculate_weight_boost_positive(self):
        """Test weight boost calculation for positive weights."""
        assert calculate_weight_boost(1.0) == 1.0
        assert calculate_weight_boost(0.5) == 0.75
        assert calculate_weight_boost(0.0) == 0.5

    def test_calculate_weight_boost_negative(self):
        """Test weight boost calculation for negative weights."""
        assert calculate_weight_boost(-1.0) == 0.0
        assert calculate_weight_boost(-0.5) == 0.25

    def test_calculate_weight_boost_clamping(self):
        """Test that weight boost clamps values outside -1 to +1."""
        assert calculate_weight_boost(2.0) == 1.0
        assert calculate_weight_boost(-2.0) == 0.0

    def test_calculate_risk_adjustment(self):
        """Test risk adjustment calculation."""
        # Low volatility should give high score
        assert calculate_risk_adjustment(0.15) == pytest.approx(1.0, abs=0.01)
        # High volatility should give low score
        assert calculate_risk_adjustment(0.50) == pytest.approx(0.0, abs=0.01)
        # None should return neutral
        assert calculate_risk_adjustment(None) == 0.5


class TestPriorityCalculator:
    """Tests for PriorityCalculator service."""

    def test_parse_industries(self):
        """Test industry string parsing."""
        assert PriorityCalculator.parse_industries("Technology") == ["Technology"]
        assert PriorityCalculator.parse_industries("Industrial, Defense") == ["Industrial", "Defense"]
        assert PriorityCalculator.parse_industries("") == []
        assert PriorityCalculator.parse_industries(None) == []

    def test_calculate_priority_basic(self):
        """Test basic priority calculation."""
        input_data = PriorityInput(
            symbol="AAPL",
            name="Apple Inc.",
            geography="US",
            industry="Technology",
            stock_score=0.7,
            volatility=0.20,
            multiplier=1.0,
            position_value=1000.0,
            total_portfolio_value=10000.0,
        )

        geo_weights = {"US": 0.2}
        industry_weights = {"Technology": 0.3}

        result = PriorityCalculator.calculate_priority(
            input_data,
            geo_weights,
            industry_weights,
        )

        assert result.symbol == "AAPL"
        assert result.combined_priority > 0
        assert result.combined_priority <= 1.0
        assert result.geo_need > 0
        assert result.industry_need > 0

    def test_calculate_priority_with_multiplier(self):
        """Test priority calculation with manual multiplier."""
        input_data = PriorityInput(
            symbol="AAPL",
            name="Apple Inc.",
            geography="US",
            industry="Technology",
            stock_score=0.6,
            volatility=0.20,
            multiplier=2.0,  # Double the priority
            position_value=0.0,
            total_portfolio_value=10000.0,
        )

        geo_weights = {"US": 0.0}
        industry_weights = {"Technology": 0.0}

        result = PriorityCalculator.calculate_priority(
            input_data,
            geo_weights,
            industry_weights,
        )

        # With multiplier 2.0, priority should be higher
        assert result.combined_priority > 0.5

    def test_calculate_priorities_sorts_by_priority(self):
        """Test that calculate_priorities sorts results by priority."""
        inputs = [
            PriorityInput(
                symbol="LOW",
                name="Low Priority",
                geography="US",
                industry="Tech",
                stock_score=0.4,
                volatility=0.30,
                multiplier=1.0,
                position_value=0.0,
                total_portfolio_value=10000.0,
            ),
            PriorityInput(
                symbol="HIGH",
                name="High Priority",
                geography="US",
                industry="Tech",
                stock_score=0.8,
                volatility=0.15,
                multiplier=1.0,
                position_value=0.0,
                total_portfolio_value=10000.0,
            ),
        ]

        geo_weights = {"US": 0.0}
        industry_weights = {"Tech": 0.0}

        results = PriorityCalculator.calculate_priorities(
            inputs,
            geo_weights,
            industry_weights,
        )

        # Should be sorted highest first
        assert results[0].symbol == "HIGH"
        assert results[1].symbol == "LOW"
        assert results[0].combined_priority > results[1].combined_priority
