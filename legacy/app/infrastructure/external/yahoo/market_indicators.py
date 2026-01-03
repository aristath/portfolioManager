"""Market indicator fetchers for forward-looking expected returns adjustments.

Fetches VIX, treasury yields, and market P/E from Yahoo Finance.
"""

import logging
from typing import Dict, Optional

from app.infrastructure.external.yahoo import data_fetchers

logger = logging.getLogger(__name__)


async def get_vix() -> Optional[float]:
    """
    Get current VIX (volatility index) level.

    Returns:
        VIX level (typically 10-50), or None if unavailable
    """
    try:
        vix_price = data_fetchers.get_current_price("^VIX")
        if vix_price and vix_price > 0:
            logger.debug(f"VIX: {vix_price:.2f}")
            return vix_price
        return None
    except Exception as e:
        logger.warning(f"Failed to get VIX: {e}")
        return None


async def get_treasury_yields() -> Dict[str, Optional[float]]:
    """
    Get current treasury yields.

    Returns:
        Dict with keys: '3m' (^IRX), '5y' (^FVX), '10y' (^TNX), '30y' (^TYX)
        Values are yields as decimals (e.g., 0.05 = 5%), or None if unavailable
    """
    symbols = {
        "3m": "^IRX",
        "5y": "^FVX",
        "10y": "^TNX",
        "30y": "^TYX",
    }

    result: Dict[str, Optional[float]] = {}
    for key, symbol in symbols.items():
        try:
            # Treasury yields are quoted as percentages, need to divide by 100
            price = data_fetchers.get_current_price(symbol)
            if price and price > 0:
                # Convert from percentage to decimal
                yield_decimal = price / 100.0
                result[key] = yield_decimal
                logger.debug(f"{key} yield: {yield_decimal*100:.2f}%")
            else:
                result[key] = None
        except Exception as e:
            logger.warning(f"Failed to get {key} yield ({symbol}): {e}")
            result[key] = None

    return result


def calculate_yield_curve_slope(yields: Dict[str, Optional[float]]) -> Optional[float]:
    """
    Calculate yield curve slope (10y - 3m).

    Positive = normal curve (expansionary)
    Negative = inverted curve (recession signal)

    Returns:
        Slope as decimal (e.g., 0.02 = 2%), or None if data unavailable
    """
    ten_y = yields.get("10y")
    three_m = yields.get("3m")

    if ten_y is not None and three_m is not None:
        slope = ten_y - three_m
        logger.debug(f"Yield curve slope: {slope*100:.2f}%")
        return slope

    return None


async def get_market_pe() -> Optional[float]:
    """
    Get market P/E ratio via SPY (S&P 500 ETF).

    Returns:
        Market P/E ratio, or None if unavailable
    """
    try:
        fundamentals = data_fetchers.get_fundamental_data("SPY.US")
        if fundamentals and fundamentals.pe_ratio:
            logger.debug(f"Market P/E (SPY): {fundamentals.pe_ratio:.2f}")
            return fundamentals.pe_ratio
        return None
    except Exception as e:
        logger.warning(f"Failed to get market P/E: {e}")
        return None
