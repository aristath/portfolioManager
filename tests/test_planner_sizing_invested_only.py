from unittest.mock import AsyncMock, MagicMock

import pytest

from sentinel.planner.planner import Planner


@pytest.mark.asyncio
async def test_planner_passes_invested_only_total_value_to_engine():
    planner = Planner()

    planner._portfolio = MagicMock()
    planner._portfolio.total_value = AsyncMock(return_value=2000.0)  # includes cash
    planner._portfolio.total_cash_eur = AsyncMock(return_value=1000.0)

    planner._portfolio_analyzer = MagicMock()
    planner._portfolio_analyzer.get_invested_value_eur = AsyncMock(return_value=1000.0)

    planner.calculate_ideal_portfolio = AsyncMock(return_value={"AAA": 1.0})
    planner.get_current_allocations = AsyncMock(return_value={"AAA": 1.0})

    planner._rebalance_engine = MagicMock()
    planner._rebalance_engine.get_recommendations = AsyncMock(return_value=[])

    await planner.get_recommendations()

    planner._rebalance_engine.get_recommendations.assert_awaited_once()
    assert planner._rebalance_engine.get_recommendations.call_args.kwargs["total_value"] == 1000.0
