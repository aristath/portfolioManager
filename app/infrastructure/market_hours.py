"""Market hours module for checking exchange trading hours.

This module provides functions to check if markets are open,
filter stocks by open markets, and group stocks by geography.
Uses the exchange_calendars library for accurate market hours
including holidays, early closes, and lunch breaks.
"""

import logging
from datetime import datetime
from functools import lru_cache
from typing import Any

import exchange_calendars as xcals
import pandas as pd
from zoneinfo import ZoneInfo

logger = logging.getLogger(__name__)

# Mapping of geography to exchange calendar codes
EXCHANGE_MAP = {
    "EU": "XETR",  # Frankfurt/XETRA
    "US": "XNYS",  # NYSE
    "ASIA": "XHKG",  # Hong Kong
}

# Timezone info for each market
MARKET_TIMEZONES = {
    "EU": "Europe/Berlin",
    "US": "America/New_York",
    "ASIA": "Asia/Hong_Kong",
}


def _get_current_time() -> datetime:
    """Get current UTC time. Extracted for testing."""
    return datetime.now(ZoneInfo("UTC"))


@lru_cache(maxsize=10)
def get_calendar(geography: str) -> Any:
    """
    Get the exchange calendar for a geography.

    Args:
        geography: One of 'EU', 'US', 'ASIA'

    Returns:
        Exchange calendar object
    """
    exchange_code = EXCHANGE_MAP.get(geography, "XNYS")
    return xcals.get_calendar(exchange_code)


def is_market_open(geography: str) -> bool:
    """
    Check if a market is currently open for trading.

    Accounts for:
    - Regular trading hours
    - Weekends
    - Holidays
    - Early closes
    - Lunch breaks (for Asian markets)

    Args:
        geography: One of 'EU', 'US', 'ASIA'

    Returns:
        True if the market is currently open
    """
    try:
        calendar = get_calendar(geography)
        now = _get_current_time()

        # Convert to pandas Timestamp in market timezone
        market_tz = calendar.tz
        now_market = pd.Timestamp(now).tz_convert(market_tz)
        today_str = now_market.strftime("%Y-%m-%d")

        # Check if today is a trading session
        if not calendar.is_session(today_str):
            return False

        # Get the schedule for today
        schedule = calendar.schedule.loc[today_str]
        open_time = schedule["open"]
        close_time = schedule["close"]

        # Check if we're within trading hours
        if not (open_time <= now_market <= close_time):
            return False

        # Check for lunch break (mainly for Asian markets)
        if "break_start" in schedule and pd.notna(schedule["break_start"]):
            break_start = schedule["break_start"]
            break_end = schedule["break_end"]
            if break_start <= now_market <= break_end:
                return False

        return True

    except Exception as e:
        logger.warning(f"Error checking market hours for {geography}: {e}")
        # Default to closed on error
        return False


def get_open_markets() -> list[str]:
    """
    Get list of currently open market geographies.

    Returns:
        List of geography strings that are currently open
    """
    return [geo for geo in EXCHANGE_MAP.keys() if is_market_open(geo)]


def get_market_status() -> dict[str, dict[str, Any]]:
    """
    Get detailed status for all markets.

    Returns:
        Dict mapping geography to status dict containing:
        - open: bool
        - exchange: str (exchange code)
        - timezone: str
        - closes_at: str (if open)
        - opens_at: str (if closed)
    """
    status = {}
    now = _get_current_time()

    for geography, exchange_code in EXCHANGE_MAP.items():
        try:
            calendar = get_calendar(geography)
            market_tz = calendar.tz
            timezone_str = MARKET_TIMEZONES.get(geography, str(market_tz))

            now_market = pd.Timestamp(now).tz_convert(market_tz)
            today_str = now_market.strftime("%Y-%m-%d")

            is_open = is_market_open(geography)

            market_info = {
                "open": is_open,
                "exchange": exchange_code,
                "timezone": timezone_str,
            }

            if is_open:
                # Get closing time for today
                schedule = calendar.schedule.loc[today_str]
                close_time = schedule["close"]
                market_info["closes_at"] = close_time.strftime("%H:%M")
            else:
                # Find next trading session
                try:
                    # Try to get next session
                    next_sessions = calendar.sessions_in_range(
                        today_str,
                        (pd.Timestamp(today_str) + pd.Timedelta(days=7)).strftime(
                            "%Y-%m-%d"
                        ),
                    )
                    if len(next_sessions) > 0:
                        # Check if today is a session but we're outside hours
                        if calendar.is_session(today_str):
                            schedule = calendar.schedule.loc[today_str]
                            if now_market < schedule["open"]:
                                # Market opens later today
                                market_info["opens_at"] = schedule["open"].strftime(
                                    "%H:%M"
                                )
                            else:
                                # Market closed for today, get next session
                                for session in next_sessions:
                                    if session.strftime("%Y-%m-%d") != today_str:
                                        next_schedule = calendar.schedule.loc[
                                            session.strftime("%Y-%m-%d")
                                        ]
                                        market_info["opens_at"] = next_schedule[
                                            "open"
                                        ].strftime("%H:%M")
                                        market_info["opens_date"] = session.strftime(
                                            "%Y-%m-%d"
                                        )
                                        break
                        else:
                            # Today is not a session, find next one
                            for session in next_sessions:
                                next_schedule = calendar.schedule.loc[
                                    session.strftime("%Y-%m-%d")
                                ]
                                market_info["opens_at"] = next_schedule[
                                    "open"
                                ].strftime("%H:%M")
                                market_info["opens_date"] = session.strftime(
                                    "%Y-%m-%d"
                                )
                                break
                except Exception:
                    market_info["opens_at"] = "Unknown"

            status[geography] = market_info

        except Exception as e:
            logger.warning(f"Error getting market status for {geography}: {e}")
            status[geography] = {
                "open": False,
                "exchange": exchange_code,
                "timezone": MARKET_TIMEZONES.get(geography, "Unknown"),
                "error": str(e),
            }

    return status


def filter_stocks_by_open_markets(stocks: list) -> list:
    """
    Filter stocks to only those whose markets are currently open.

    Args:
        stocks: List of stock objects with 'geography' attribute

    Returns:
        Filtered list of stocks with open markets
    """
    open_markets = get_open_markets()
    return [s for s in stocks if getattr(s, "geography", None) in open_markets]


def group_stocks_by_geography(stocks: list) -> dict[str, list]:
    """
    Group stocks by their geography.

    Args:
        stocks: List of stock objects with 'geography' attribute

    Returns:
        Dict mapping geography to list of stocks
    """
    grouped: dict[str, list] = {"EU": [], "US": [], "ASIA": []}

    for stock in stocks:
        geography = getattr(stock, "geography", None)
        if geography in grouped:
            grouped[geography].append(stock)

    return grouped
