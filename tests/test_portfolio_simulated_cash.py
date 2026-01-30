"""Tests for Portfolio simulated cash behavior in research mode.

When trading_mode is 'research' and simulated_cash_eur is set,
Portfolio should return the simulated value instead of real broker data.
"""

from unittest.mock import AsyncMock, MagicMock

import pytest

from sentinel.portfolio import Portfolio


@pytest.fixture
def mock_db():
    db = MagicMock()
    db.get_cash_balances = AsyncMock(return_value={"EUR": 1000.0, "USD": 500.0})
    return db


@pytest.fixture
def mock_settings():
    settings = MagicMock()
    settings.get = AsyncMock(
        side_effect=lambda key, **kw: {
            "trading_mode": "research",
            "simulated_cash_eur": 5000.0,
        }.get(key, kw.get("default"))
    )
    return settings


@pytest.fixture
def mock_currency():
    currency = MagicMock()
    currency.to_eur = AsyncMock(side_effect=lambda amt, curr: amt * 0.85 if curr == "USD" else amt)
    return currency


class TestSimulatedCash:
    """Tests for simulated cash in research mode."""

    @pytest.mark.asyncio
    async def test_total_cash_eur_returns_simulated_in_research_mode(self, mock_db, mock_settings, mock_currency):
        """Simulated value returned when research mode + setting set."""
        portfolio = Portfolio(db=mock_db)
        portfolio._settings = mock_settings
        portfolio._currency = mock_currency

        result = await portfolio.total_cash_eur()
        assert result == 5000.0

    @pytest.mark.asyncio
    async def test_total_cash_eur_returns_real_when_simulated_not_set(self, mock_db, mock_currency):
        """Real value returned when simulated_cash_eur is None."""
        settings = MagicMock()
        settings.get = AsyncMock(
            side_effect=lambda key, **kw: {
                "trading_mode": "research",
                "simulated_cash_eur": None,
            }.get(key, kw.get("default"))
        )

        portfolio = Portfolio(db=mock_db)
        portfolio._settings = settings
        portfolio._currency = mock_currency

        result = await portfolio.total_cash_eur()
        # Real: EUR 1000 + USD 500 * 0.85 = 1425.0
        assert result == 1425.0

    @pytest.mark.asyncio
    async def test_total_cash_eur_returns_real_in_live_mode(self, mock_db, mock_currency):
        """Real value returned in live mode even if simulated_cash_eur is set."""
        settings = MagicMock()
        settings.get = AsyncMock(
            side_effect=lambda key, **kw: {
                "trading_mode": "live",
                "simulated_cash_eur": 5000.0,
            }.get(key, kw.get("default"))
        )

        portfolio = Portfolio(db=mock_db)
        portfolio._settings = settings
        portfolio._currency = mock_currency

        result = await portfolio.total_cash_eur()
        # Real: EUR 1000 + USD 500 * 0.85 = 1425.0
        assert result == 1425.0

    @pytest.mark.asyncio
    async def test_get_cash_balances_returns_eur_dict_when_simulated(self, mock_db, mock_settings):
        """Returns {'EUR': value} dict in research mode with simulated cash."""
        portfolio = Portfolio(db=mock_db)
        portfolio._settings = mock_settings

        result = await portfolio.get_cash_balances()
        assert result == {"EUR": 5000.0}

    @pytest.mark.asyncio
    async def test_get_cash_balances_returns_real_when_not_simulated(self, mock_db, mock_currency):
        """Returns normal multi-currency dict when not simulated."""
        settings = MagicMock()
        settings.get = AsyncMock(
            side_effect=lambda key, **kw: {
                "trading_mode": "live",
                "simulated_cash_eur": None,
            }.get(key, kw.get("default"))
        )

        portfolio = Portfolio(db=mock_db)
        portfolio._settings = settings
        portfolio._currency = mock_currency

        result = await portfolio.get_cash_balances()
        assert result == {"EUR": 1000.0, "USD": 500.0}

    @pytest.mark.asyncio
    async def test_invalid_simulated_value_falls_back_to_real(self, mock_db, mock_currency):
        """Invalid (non-numeric) simulated value falls back to real cash."""
        settings = MagicMock()
        settings.get = AsyncMock(
            side_effect=lambda key, **kw: {
                "trading_mode": "research",
                "simulated_cash_eur": "",
            }.get(key, kw.get("default"))
        )

        portfolio = Portfolio(db=mock_db)
        portfolio._settings = settings
        portfolio._currency = mock_currency

        result = await portfolio.total_cash_eur()
        # Falls back to real: EUR 1000 + USD 500 * 0.85 = 1425.0
        assert result == 1425.0
