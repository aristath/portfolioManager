"""Tradernet API client package.

This package provides the Tradernet (Freedom24) API client and related utilities.

The package is organized into modules:
- models.py: Data classes (Position, CashBalance, Quote, OHLC, OrderResult)
- parsers.py: Response parsing functions
- transactions.py: Transaction parsing and processing functions
- utils.py: Shared utilities (LED API call context, exchange rate sync)
- client.py: Main TradernetClient class and singleton getter
"""

from typing import Optional

from tradernet import TraderNetAPI

from app.infrastructure.external.tradernet.client import (
    TradernetClient,
    get_tradernet_client,
)
from app.infrastructure.external.tradernet.models import (
    OHLC,
    CashBalance,
    OrderResult,
    Position,
    Quote,
)
from app.infrastructure.external.tradernet.utils import get_exchange_rate_sync

# Alias for backward compatibility
Tradernet = TraderNetAPI


# Backward compatibility - deprecated function
def get_exchange_rate(from_currency: str, to_currency: Optional[str] = None) -> float:
    """Get exchange rate from currency to target currency (deprecated).

    This function is kept for backward compatibility but will be removed.
    Use ExchangeRateService directly in async code, or get_exchange_rate_sync() in sync code.
    """
    from app.domain.value_objects.currency import Currency

    if to_currency is None:
        to_currency = Currency.EUR

    return get_exchange_rate_sync(from_currency, to_currency)


__all__ = [
    "TradernetClient",
    "get_tradernet_client",
    "Tradernet",
    "TraderNetAPI",
    "Position",
    "CashBalance",
    "Quote",
    "OHLC",
    "OrderResult",
    "get_exchange_rate",  # Deprecated but exported for backward compatibility
]
