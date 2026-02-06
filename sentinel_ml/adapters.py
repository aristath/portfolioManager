"""Adapters bridging ML domain code to monolith HTTP data APIs."""

from __future__ import annotations

from sentinel_ml.clients.monolith_client import MonolithDataClient


class MonolithDBAdapter:
    """Small async interface compatible with ML domain components."""

    def __init__(self, client: MonolithDataClient):
        self.client = client

    async def connect(self):
        return self

    async def get_all_securities(self, active_only: bool = True):
        return await self.client.get_securities(active_only=active_only, ml_enabled_only=False)

    async def get_ml_enabled_securities(self):
        return await self.client.get_ml_enabled_securities()

    async def get_security(self, symbol: str):
        return await self.client.get_security(symbol)

    async def get_prices(self, symbol: str, days: int = 3650, end_date: str | None = None):
        return await self.client.get_prices(symbol, days=days, end_date=end_date)

    async def get_prices_bulk(self, symbols: list[str]):
        result: dict[str, list[dict]] = {}
        for symbol in symbols:
            result[symbol] = await self.get_prices(symbol, days=3650)
        return result


class MonolithSettingsAdapter:
    def __init__(self, client: MonolithDataClient):
        self.client = client

    async def get(self, key: str, default=None):
        values = await self.client.get_settings([key])
        value = values.get(key)
        return default if value is None else value
