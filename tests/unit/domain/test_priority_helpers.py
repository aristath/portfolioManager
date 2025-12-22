"""Unit tests for priority helper functions."""

import pytest

from app.domain.utils.priority_helpers import (
    calculate_weight_boost,
    calculate_risk_adjustment,
)


class TestCalculateWeightBoost:
    """Tests for calculate_weight_boost function."""

    def test_positive_weights(self):
        """Test positive weight values."""
        assert calculate_weight_boost(1.0) == 1.0
        assert calculate_weight_boost(0.5) == 0.75
        assert calculate_weight_boost(0.0) == 0.5

    def test_negative_weights(self):
        """Test negative weight values."""
        assert calculate_weight_boost(-1.0) == 0.0
        assert calculate_weight_boost(-0.5) == 0.25

    def test_clamping(self):
        """Test that values outside range are clamped."""
        assert calculate_weight_boost(2.0) == 1.0
        assert calculate_weight_boost(-2.0) == 0.0


class TestCalculateRiskAdjustment:
    """Tests for calculate_risk_adjustment function."""

    def test_low_volatility(self):
        """Test that low volatility gives high score."""
        assert calculate_risk_adjustment(0.15) == pytest.approx(1.0, abs=0.01)

    def test_high_volatility(self):
        """Test that high volatility gives low score."""
        assert calculate_risk_adjustment(0.50) == pytest.approx(0.0, abs=0.01)

    def test_medium_volatility(self):
        """Test medium volatility values."""
        result = calculate_risk_adjustment(0.30)
        assert 0.0 < result < 1.0

    def test_none_volatility(self):
        """Test that None volatility returns neutral."""
        assert calculate_risk_adjustment(None) == 0.5
