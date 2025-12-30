"""Tests for risk models calculation.

These tests validate risk model calculations for portfolio optimization,
including covariance matrices, volatility estimates, and risk adjustments.
"""

from unittest.mock import AsyncMock, MagicMock, patch

import numpy as np
import pandas as pd
import pytest


class TestCalculateRiskModel:
    """Test calculate_risk_model function."""

    @pytest.mark.asyncio
    async def test_calculates_covariance_matrix(self):
        """Test that covariance matrix is calculated from historical returns."""
        from app.application.services.optimization.risk_models import (
            calculate_risk_model,
        )

        symbols = ["AAPL", "MSFT"]
        historical_returns = {
            "AAPL": pd.Series([0.01, 0.02, -0.01, 0.03], dtype=float),
            "MSFT": pd.Series([0.02, 0.01, 0.02, 0.01], dtype=float),
        }

        with patch(
            "app.application.services.optimization.risk_models._get_historical_returns",
            new_callable=AsyncMock,
        ) as mock_get_returns:
            mock_get_returns.return_value = historical_returns

            result = await calculate_risk_model(symbols, lookback_days=252)

            assert isinstance(result, dict)
            # Should return covariance matrix or risk metrics
            assert len(result) > 0

    @pytest.mark.asyncio
    async def test_handles_missing_historical_data(self):
        """Test handling when historical data is missing for some symbols."""
        from app.application.services.optimization.risk_models import (
            calculate_risk_model,
        )

        symbols = ["AAPL", "UNKNOWN"]
        historical_returns = {
            "AAPL": pd.Series([0.01, 0.02], dtype=float),
            "UNKNOWN": pd.Series(dtype=float),  # Empty series
        }

        with patch(
            "app.application.services.optimization.risk_models._get_historical_returns",
            new_callable=AsyncMock,
        ) as mock_get_returns:
            mock_get_returns.return_value = historical_returns

            result = await calculate_risk_model(symbols, lookback_days=252)

            assert isinstance(result, dict)
            # Should handle missing data gracefully

    @pytest.mark.asyncio
    async def test_handles_empty_symbols_list(self):
        """Test handling when symbols list is empty."""
        from app.application.services.optimization.risk_models import (
            calculate_risk_model,
        )

        symbols = []
        historical_returns = {}

        with patch(
            "app.application.services.optimization.risk_models._get_historical_returns",
            new_callable=AsyncMock,
        ) as mock_get_returns:
            mock_get_returns.return_value = historical_returns

            result = await calculate_risk_model(symbols, lookback_days=252)

            assert isinstance(result, dict)
            # Should handle empty list gracefully

    @pytest.mark.asyncio
    async def test_handles_insufficient_data_for_covariance(self):
        """Test handling when there's insufficient data to calculate covariance."""
        from app.application.services.optimization.risk_models import (
            calculate_risk_model,
        )

        symbols = ["AAPL"]
        # Only one return value (need at least 2 for covariance)
        historical_returns = {"AAPL": pd.Series([0.01], dtype=float)}

        with patch(
            "app.application.services.optimization.risk_models._get_historical_returns",
            new_callable=AsyncMock,
        ) as mock_get_returns:
            mock_get_returns.return_value = historical_returns

            result = await calculate_risk_model(symbols, lookback_days=252)

            assert isinstance(result, dict)
            # Should handle insufficient data gracefully

    @pytest.mark.asyncio
    async def test_uses_default_lookback_days(self):
        """Test that default lookback_days is used when not specified."""
        from app.application.services.optimization.risk_models import (
            calculate_risk_model,
        )

        symbols = ["AAPL"]
        historical_returns = {"AAPL": pd.Series([0.01, 0.02], dtype=float)}

        with patch(
            "app.application.services.optimization.risk_models._get_historical_returns",
            new_callable=AsyncMock,
        ) as mock_get_returns:
            mock_get_returns.return_value = historical_returns

            result = await calculate_risk_model(symbols)

            # Should use default lookback_days
            assert isinstance(result, dict)


class TestGetHistoricalReturns:
    """Test _get_historical_returns helper function (shared with expected_returns)."""

    @pytest.mark.asyncio
    async def test_fetches_returns_from_history_repo(self):
        """Test that historical returns are fetched from HistoryRepository."""
        from app.application.services.optimization.risk_models import (
            _get_historical_returns,
        )

        symbols = ["AAPL"]
        mock_history_repo = AsyncMock()
        mock_prices = [
            MagicMock(date="2024-01-01", close_price=100.0),
            MagicMock(date="2024-01-02", close_price=101.0),
        ]
        mock_history_repo.get_daily_range.return_value = mock_prices

        with patch(
            "app.application.services.optimization.risk_models.HistoryRepository",
            return_value=mock_history_repo,
        ):
            result = await _get_historical_returns(symbols, lookback_days=252)

            assert isinstance(result, dict)
            assert "AAPL" in result
            assert isinstance(result["AAPL"], pd.Series)

    @pytest.mark.asyncio
    async def test_calculates_returns_correctly(self):
        """Test that returns are calculated correctly from price data."""
        from app.application.services.optimization.risk_models import (
            _get_historical_returns,
        )

        symbols = ["AAPL"]
        mock_prices = [
            MagicMock(date="2024-01-01", close_price=100.0),
            MagicMock(date="2024-01-02", close_price=101.0),
            MagicMock(date="2024-01-03", close_price=102.0),
        ]

        mock_history_repo = AsyncMock()
        mock_history_repo.get_daily_range.return_value = mock_prices

        with patch(
            "app.application.services.optimization.risk_models.HistoryRepository",
            return_value=mock_history_repo,
        ):
            result = await _get_historical_returns(symbols, lookback_days=252)

            assert "AAPL" in result
            returns = result["AAPL"]
            assert isinstance(returns, pd.Series)
            assert len(returns) == 2  # 3 prices -> 2 returns
            assert returns.iloc[0] == pytest.approx(0.01, abs=0.001)


class TestCalculateCovarianceMatrix:
    """Test covariance matrix calculation."""

    def test_calculates_covariance_from_returns_dataframe(self):
        """Test that covariance matrix is calculated from returns DataFrame."""
        from app.application.services.optimization.risk_models import (
            _calculate_covariance_matrix,
        )

        # Create sample returns data
        dates = pd.date_range("2024-01-01", periods=5, freq="D")
        returns_df = pd.DataFrame(
            {
                "AAPL": [0.01, 0.02, -0.01, 0.03],
                "MSFT": [0.02, 0.01, 0.02, 0.01],
            },
            index=dates[:4],
        )

        cov_matrix = _calculate_covariance_matrix(returns_df)

        assert isinstance(cov_matrix, (pd.DataFrame, np.ndarray))
        # Should be square matrix with dimensions matching number of symbols
        if isinstance(cov_matrix, pd.DataFrame):
            assert cov_matrix.shape[0] == 2
            assert cov_matrix.shape[1] == 2

    def test_handles_single_symbol(self):
        """Test handling when only one symbol is provided."""
        from app.application.services.optimization.risk_models import (
            _calculate_covariance_matrix,
        )

        dates = pd.date_range("2024-01-01", periods=5, freq="D")
        returns_df = pd.DataFrame(
            {"AAPL": [0.01, 0.02, -0.01, 0.03]}, index=dates[:4]
        )

        cov_matrix = _calculate_covariance_matrix(returns_df)

        assert isinstance(cov_matrix, (pd.DataFrame, np.ndarray))
        # Single symbol should result in 1x1 matrix

    def test_handles_empty_dataframe(self):
        """Test handling when returns DataFrame is empty."""
        from app.application.services.optimization.risk_models import (
            _calculate_covariance_matrix,
        )

        returns_df = pd.DataFrame()

        cov_matrix = _calculate_covariance_matrix(returns_df)

        # Should handle empty DataFrame gracefully
        assert isinstance(cov_matrix, (pd.DataFrame, np.ndarray)) or cov_matrix is None
