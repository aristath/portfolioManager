"""Satellites domain enumerations."""

from enum import Enum


class BucketType(str, Enum):
    """Type of portfolio bucket."""

    CORE = "core"
    SATELLITE = "satellite"


class BucketStatus(str, Enum):
    """Status of a portfolio bucket.

    Lifecycle:
    - research: Paper trading, no real money
    - accumulating: Building up to minimum threshold from deposits
    - active: Normal trading
    - hibernating: Below minimum, hold only, no new trades
    - paused: Manual pause by user
    - retired: Closed, historical data only
    """

    RESEARCH = "research"
    ACCUMULATING = "accumulating"
    ACTIVE = "active"
    HIBERNATING = "hibernating"
    PAUSED = "paused"
    RETIRED = "retired"


class TransactionType(str, Enum):
    """Type of bucket transaction for audit trail."""

    DEPOSIT = "deposit"
    REALLOCATION = "reallocation"
    TRADE_BUY = "trade_buy"
    TRADE_SELL = "trade_sell"
    DIVIDEND = "dividend"
    TRANSFER_IN = "transfer_in"
    TRANSFER_OUT = "transfer_out"
