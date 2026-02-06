"""Tests for securities API endpoints."""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest


@pytest.mark.asyncio
async def test_get_unified_view_returns_empty_list_when_no_securities():
    """GET /api/unified returns empty list when no securities exist."""
    from sentinel.api.routers.securities import get_unified_view

    # Create mock dependencies
    mock_deps = MagicMock()
    mock_deps.db.get_all_securities = AsyncMock(return_value=[])

    # Call the endpoint
    result = await get_unified_view(mock_deps, period="1Y")

    # Verify it returns an empty list
    assert result == []

    # Verify get_all_securities was called
    mock_deps.db.get_all_securities.assert_called_once_with(active_only=True)


@pytest.mark.asyncio
async def test_get_unified_view_does_not_call_planner_when_no_securities():
    """Verify planner is not instantiated when securities list is empty."""
    from sentinel.api.routers.securities import get_unified_view

    # Create mock dependencies
    mock_deps = MagicMock()
    mock_deps.db.get_all_securities = AsyncMock(return_value=[])

    # Mock Planner where it is defined (router imports it inside the endpoint)
    with patch("sentinel.planner.Planner") as mock_planner:
        result = await get_unified_view(mock_deps, period="1Y")

    # Verify Planner was not instantiated
    mock_planner.assert_not_called()
    assert result == []


def _make_unified_mocks(one_security=True):
    """Build mocks so get_unified_view runs without hitting real deps."""
    mock_deps = MagicMock()
    if one_security:
        mock_deps.db.get_all_securities = AsyncMock(
            return_value=[{"symbol": "AAPL", "name": "Apple", "currency": "USD"}]
        )
    else:
        mock_deps.db.get_all_securities = AsyncMock(return_value=[])
    mock_deps.db.get_all_positions = AsyncMock(return_value=[])
    mock_cursor = MagicMock()
    mock_cursor.fetchall = AsyncMock(return_value=[])
    mock_cursor.fetchone = AsyncMock(return_value=None)
    mock_deps.db.conn.execute = AsyncMock(return_value=mock_cursor)
    mock_deps.broker.get_quotes = AsyncMock(return_value={})
    mock_deps.db.get_prices_bulk = AsyncMock(return_value={"AAPL": []})
    mock_deps.currency.to_eur = AsyncMock(return_value=0.0)

    # ML db mocks for per-model predictions
    mock_deps.settings.get = AsyncMock(return_value=0.25)
    ml_cursor = MagicMock()
    ml_cursor.fetchone = AsyncMock(return_value=None)
    ml_cursor.fetchall = AsyncMock(return_value=[])
    mock_deps.ml_db.conn.execute = AsyncMock(return_value=ml_cursor)
    mock_deps.ml_db.get_prediction_as_of = AsyncMock(return_value=None)
    return mock_deps


@pytest.mark.asyncio
async def test_get_unified_view_with_as_of_reads_per_model_predictions():
    """When as_of is set, endpoint reads per-model predictions from ml_db."""
    from sentinel.api.routers.securities import get_unified_view

    mock_deps = _make_unified_mocks(one_security=True)
    mock_planner = MagicMock()
    mock_planner.get_recommendations = AsyncMock(return_value=[])
    mock_planner.calculate_ideal_portfolio = AsyncMock(return_value={})
    mock_planner.get_current_allocations = AsyncMock(return_value={})

    with patch("sentinel.planner.Planner", return_value=mock_planner):
        await get_unified_view(mock_deps, period="1Y", as_of="2024-01-15")

    # Should read from ml_db per-model tables
    mock_deps.ml_db.get_prediction_as_of.assert_called()


@pytest.mark.asyncio
async def test_get_unified_view_without_as_of_reads_latest_per_model():
    """When as_of is None, endpoint reads latest per-model predictions from ml_db."""
    from sentinel.api.routers.securities import get_unified_view

    mock_deps = _make_unified_mocks(one_security=True)
    mock_planner = MagicMock()
    mock_planner.get_recommendations = AsyncMock(return_value=[])
    mock_planner.calculate_ideal_portfolio = AsyncMock(return_value={})
    mock_planner.get_current_allocations = AsyncMock(return_value={})

    with patch("sentinel.planner.Planner", return_value=mock_planner):
        await get_unified_view(mock_deps, period="1Y", as_of=None)

    # Should query ml_db for latest predictions (no as_of)
    ml_execute_calls = [str(c) for c in mock_deps.ml_db.conn.execute.call_args_list]
    assert any("ml_predictions_" in c for c in ml_execute_calls)
