"""Tests for Analyzer (scoring, analyze_prices)."""

import pytest

from sentinel.analyzer import Analyzer


class TestAnalyzePrices:
    """Tests for analyze_prices (prices in, score out, no regime)."""

    @pytest.mark.asyncio
    async def test_analyze_prices_returns_none_for_insufficient_data(self):
        """analyze_prices returns None when fewer than 252 points."""
        analyzer = Analyzer()
        prices = [{"close": 100.0 + i} for i in range(100)]
        result = await analyzer.analyze_prices("SYM", prices)
        assert result is None

    @pytest.mark.asyncio
    async def test_analyze_prices_returns_score_and_components(self):
        """analyze_prices returns (score, components) with expected keys when enough data."""
        analyzer = Analyzer()
        # 252 trading days of synthetic data (slight upward trend)
        prices = [{"close": 100.0 + (i * 0.1) + (i % 10)} for i in range(252)]
        result = await analyzer.analyze_prices("SYM", prices)
        assert result is not None
        score, components = result
        assert isinstance(score, float)
        expected_keys = {
            "trend",
            "trend_strength",
            "momentum",
            "cycle_position",
            "cycle_position_pct",
            "expected_return",
            "sharpe",
            "cagr",
            "consistency",
            "max_drawdown",
            "volatility",
            "volatility_trend",
            "long_term_component",
            "medium_term_component",
            "short_term_component",
            "noise_component",
        }
        assert expected_keys.issubset(components.keys())
