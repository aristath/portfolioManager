"""Yahoo Finance service for analyst data and fundamentals.

This module provides a unified interface to Yahoo Finance data.
All functionality is implemented in sub-modules for better organization.
"""

# Re-export all public functions and classes for backward compatibility
from app.infrastructure.external.yahoo.symbol_converter import get_yahoo_symbol
from app.infrastructure.external.yahoo.models import (
    AnalystData,
    FundamentalData,
    HistoricalPrice,
)
from app.infrastructure.external.yahoo.data_fetchers import (
    get_analyst_data,
    get_fundamental_data,
    get_historical_prices,
    get_current_price,
    get_stock_industry,
    get_batch_quotes,
)

__all__ = [
    "get_yahoo_symbol",
    "AnalystData",
    "FundamentalData",
    "HistoricalPrice",
    "get_analyst_data",
    "get_fundamental_data",
    "get_historical_prices",
    "get_current_price",
    "get_stock_industry",
    "get_batch_quotes",
]
