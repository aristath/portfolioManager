from unittest.mock import AsyncMock

import pytest

from sentinel.planner.analyzer import PortfolioAnalyzer


@pytest.mark.asyncio
async def test_current_allocations_exclude_cash_from_denominator():
    analyzer = PortfolioAnalyzer()

    analyzer._db = type("Db", (), {})()
    analyzer._db.cache_get = AsyncMock(return_value=None)
    analyzer._db.cache_set = AsyncMock(return_value=None)

    analyzer._portfolio = type("Portfolio", (), {})()
    analyzer._portfolio.positions = AsyncMock(
        return_value=[
            {"symbol": "AAA", "quantity": 10, "current_price": 100, "currency": "EUR"},
        ]
    )
    analyzer._portfolio.total_value = AsyncMock(return_value=2000.0)  # includes cash

    analyzer._currency = type("Currency", (), {})()
    analyzer._currency.get_rate = AsyncMock(return_value=1.0)

    allocations = await analyzer.get_current_allocations()

    assert allocations == {"AAA": 1.0}
