"""Tests for job task functions."""

from unittest.mock import AsyncMock, MagicMock

import pytest

from sentinel.jobs.tasks import sync_prices


class TestSyncPricesClearsFeatureCache:
    """Tests that sync_prices clears feature cache before fetching new prices."""

    @pytest.mark.asyncio
    async def test_sync_prices_clears_feature_cache_before_fetch(self):
        """Verify cache_clear("features:") is called before broker.get_historical_prices_bulk."""
        # Track cross-object call ordering via a shared list
        call_order = []

        db = MagicMock()
        db.get_all_securities = AsyncMock(
            return_value=[
                {"symbol": "TEST.EU"},
            ]
        )

        async def track_cache_clear(prefix):
            call_order.append(("db.cache_clear", prefix))
            return 1

        db.cache_clear = AsyncMock(side_effect=track_cache_clear)
        db.save_prices = AsyncMock()

        broker = MagicMock()

        async def track_fetch(symbols, **kwargs):
            call_order.append(("broker.get_historical_prices_bulk",))
            return {"TEST.EU": [{"date": "2025-01-01", "close": 100.0}]}

        broker.get_historical_prices_bulk = AsyncMock(side_effect=track_fetch)

        cache = MagicMock()
        cache.clear = MagicMock(return_value=5)

        await sync_prices(db, broker, cache)

        # cache_clear("features:") must have been called
        db.cache_clear.assert_called_once_with("features:")

        # Verify ordering: cache_clear happens before broker fetch
        clear_idx = next(i for i, c in enumerate(call_order) if c[0] == "db.cache_clear")
        fetch_idx = next(i for i, c in enumerate(call_order) if c[0] == "broker.get_historical_prices_bulk")
        assert clear_idx < fetch_idx, (
            f"cache_clear (index {clear_idx}) must precede get_historical_prices_bulk (index {fetch_idx})"
        )
