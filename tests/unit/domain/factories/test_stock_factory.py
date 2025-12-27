"""Tests for StockFactory."""

import pytest

from app.domain.exceptions import ValidationError
from app.domain.factories.stock_factory import StockFactory
from app.domain.value_objects.currency import Currency


class TestStockFactory:
    """Test StockFactory creation methods."""

    def test_create_from_api_request_valid(self):
        """Test creating stock from valid API request."""
        data = {
            "symbol": "AAPL.US",
            "name": "Apple Inc.",
            "country": "United States",
            "industry": "Consumer Electronics",
            "min_lot": 1,
            "allow_buy": True,
            "allow_sell": False,
        }

        stock = StockFactory.create_from_api_request(data)

        assert stock.symbol == "AAPL.US"
        assert stock.name == "Apple Inc."
        assert stock.country == "United States"
        assert stock.industry == "Consumer Electronics"
        assert stock.min_lot == 1
        assert stock.allow_buy is True
        assert stock.allow_sell is False
        assert stock.active is True
        assert stock.priority_multiplier == 1.0

    def test_create_from_api_request_normalizes_symbol(self):
        """Test that symbol is normalized to uppercase."""
        data = {
            "symbol": "aapl.us",
            "name": "Apple Inc.",
            "country": "United States",
        }

        stock = StockFactory.create_from_api_request(data)
        assert stock.symbol == "AAPL.US"

    def test_create_from_api_request_accepts_country(self):
        """Test that country is accepted."""
        data = {
            "symbol": "AAPL.US",
            "name": "Apple Inc.",
            "country": "United States",
        }

        stock = StockFactory.create_from_api_request(data)
        assert stock.country == "United States"

    def test_create_from_api_request_sets_currency_from_country(self):
        """Test that currency is set based on country (currency comes from stock data, not country)."""
        # Currency is now synced from Tradernet, not derived from country
        data = {
            "symbol": "AAPL.US",
            "name": "Apple Inc.",
            "country": "United States",
            "currency": "USD",
        }
        stock = StockFactory.create_from_api_request(data)
        assert stock.currency == Currency.USD

        data = {
            "symbol": "SAP.DE",
            "name": "SAP SE",
            "country": "Germany",
            "currency": "EUR",
        }
        stock = StockFactory.create_from_api_request(data)
        assert stock.currency == Currency.EUR

    def test_create_from_api_request_validates_min_lot(self):
        """Test that min_lot is validated (must be >= 1)."""
        data = {
            "symbol": "AAPL.US",
            "name": "Apple Inc.",
            "country": "United States",
            "min_lot": 0,  # Invalid
        }

        stock = StockFactory.create_from_api_request(data)
        assert stock.min_lot == 1  # Should default to 1

        data["min_lot"] = -5
        stock = StockFactory.create_from_api_request(data)
        assert stock.min_lot == 1

    def test_create_from_api_request_validates_symbol_not_empty(self):
        """Test that symbol cannot be empty."""
        data = {
            "symbol": "",
            "name": "Apple Inc.",
            "country": "United States",
        }

        with pytest.raises(ValidationError, match="Symbol cannot be empty"):
            StockFactory.create_from_api_request(data)

    def test_create_from_api_request_accepts_any_country(self):
        """Test that any non-empty country is accepted."""
        data = {
            "symbol": "AAPL.US",
            "name": "Apple Inc.",
            "country": "Greece",
        }

        stock = StockFactory.create_from_api_request(data)
        assert stock.country == "Greece"

    def test_create_with_industry_detection(self):
        """Test creating stock with industry detection."""
        # Mock industry detection would go here
        # For now, test that it accepts industry parameter
        data = {
            "symbol": "AAPL.US",
            "name": "Apple Inc.",
            "country": "United States",
            "industry": "Consumer Electronics",
        }

        stock = StockFactory.create_from_api_request(data)
        assert stock.industry == "Consumer Electronics"

    def test_create_from_import(self):
        """Test creating stock from import data."""
        data = {
            "symbol": "MSFT.US",
            "name": "Microsoft Corporation",
            "country": "United States",
            "industry": "Consumer Electronics",
            "yahoo_symbol": "MSFT",
            "currency": "USD",
        }

        stock = StockFactory.create_from_import(data)

        assert stock.symbol == "MSFT.US"
        assert stock.name == "Microsoft Corporation"
        assert stock.country == "United States"
        assert stock.industry == "Consumer Electronics"
        assert stock.yahoo_symbol == "MSFT"
        assert stock.currency == Currency.USD
