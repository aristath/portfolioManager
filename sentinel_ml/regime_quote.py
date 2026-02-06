"""
Regime detection and ML prediction dampening.

Uses real-time quote data to detect per-security market regime
and dampen ML predictions when regime disagrees with prediction direction.
"""

import json
import logging
from typing import Any, Tuple, Union

import pandas as pd

logger = logging.getLogger(__name__)

WEEK52_TRADING_DAYS = 252


def quote_data_from_prices(
    prices: Union[list[dict], pd.DataFrame],
) -> dict:
    """
    Build quote_data-shaped dict from historical OHLCV (newest first).

    Used for historical regime in backfill when live quote_data is not available.
    Input rows must be ordered newest first (as returned by get_prices).

    Args:
        prices: List of dicts with at least "date", "close"; optional "high", "low".
                Or DataFrame with same columns.

    Returns:
        Dict with chg5, chg22, chg110, chg220, ltp, x_max, x_min (same shape as
        Tradernet quote_data used by calculate_regime_score).
    """
    if isinstance(prices, pd.DataFrame):
        prices = prices.to_dict("records")
    if not prices:
        return {
            "chg5": 0,
            "chg22": 0,
            "chg110": 0,
            "chg220": 0,
            "ltp": 0,
            "x_max": 0,
            "x_min": 0,
        }
    # Newest first: row 0 = most recent
    close_series = [float(p.get("close") or 0) for p in prices]
    n = len(close_series)
    ltp = close_series[0]

    # 52-week high/low: use high/low if present, else close
    window = min(n, WEEK52_TRADING_DAYS)
    highs = [float(p.get("high") or p.get("close") or 0) for p in prices[:window]]
    lows = [float(p.get("low") or p.get("close") or 0) for p in prices[:window]]
    if not highs:
        x_max = ltp
        x_min = ltp
    else:
        x_max = max(highs)
        x_min = min(lows)

    def pct_chg(ago: int) -> float:
        if ago >= n or close_series[ago] == 0:
            return 0.0
        return (ltp - close_series[ago]) / close_series[ago] * 100

    chg5 = pct_chg(5) if n > 5 else 0.0
    chg22 = pct_chg(22) if n > 22 else 0.0
    chg110 = pct_chg(110) if n > 110 else 0.0
    chg220 = pct_chg(220) if n > 220 else 0.0

    return {
        "chg5": chg5,
        "chg22": chg22,
        "chg110": chg110,
        "chg220": chg220,
        "ltp": ltp,
        "x_max": x_max,
        "x_min": x_min,
    }


def calculate_regime_score(quote_data: dict) -> float:
    """
    Calculate continuous regime score from -1 (bearish) to +1 (bullish).

    Args:
        quote_data: Raw Tradernet quote response with chg5, chg22, etc.

    Returns:
        Regime score between -1.0 and +1.0
    """
    chg5 = quote_data.get("chg5", 0) or 0
    chg22 = quote_data.get("chg22", 0) or 0
    chg110 = quote_data.get("chg110", 0) or 0
    chg220 = quote_data.get("chg220", 0) or 0

    # Multi-timeframe momentum (weighted toward medium-term)
    # Normalize each timeframe by typical range
    momentum = (
        0.10 * (chg5 / 10)  # 5-day, +/-10% typical
        + 0.25 * (chg22 / 20)  # 1-month, +/-20% typical
        + 0.40 * (chg110 / 40)  # ~6-month, heaviest weight
        + 0.25 * (chg220 / 60)  # ~1-year, +/-60% typical
    )
    momentum = max(-1.0, min(1.0, momentum))

    # Position in 52-week range (0 = at lows, 1 = at highs)
    ltp = quote_data.get("ltp", 0) or 0
    x_max = quote_data.get("x_max", 0) or 0
    x_min = quote_data.get("x_min", 0) or 0

    if x_max > x_min and ltp > 0:
        position = (ltp - x_min) / (x_max - x_min)
    else:
        position = 0.5  # Neutral if data missing

    # Combine: 70% momentum, 30% position-adjusted
    regime_score = momentum * 0.7 + (position - 0.5) * 0.6
    return max(-1.0, min(1.0, regime_score))


def apply_regime_dampening(ml_return: float, regime_score: float, max_dampening: float = 0.4) -> float:
    """
    Apply regime-based dampening to ML return prediction.

    Only dampens when regime disagrees with prediction direction.

    Args:
        ml_return: ML predicted return (e.g., 0.05 for 5%)
        regime_score: Regime score from -1 to +1
        max_dampening: Maximum reduction factor (default 0.4 = 40%)

    Returns:
        Adjusted return prediction
    """
    if ml_return > 0 and regime_score < 0:
        # ML bullish, regime bearish
        disagreement = abs(regime_score)
        dampening = disagreement * max_dampening
    elif ml_return < 0 and regime_score > 0:
        # ML bearish, regime bullish
        disagreement = regime_score
        dampening = disagreement * max_dampening
    else:
        dampening = 0

    return ml_return * (1 - dampening)


async def get_regime_adjusted_return(
    symbol: str,
    ml_return: float,
    db: Any,
    quote_data: dict | None = None,
) -> Tuple[float, float, float]:
    """
    Get regime-adjusted return for a symbol.

    Args:
        symbol: Security symbol
        ml_return: Raw ML prediction
        db: Database instance (used only when quote_data is None)
        quote_data: If provided, use this instead of loading from DB (no db.get_security call).

    Returns:
        Tuple of (adjusted_return, regime_score, dampening_applied)
    """
    if quote_data is not None:
        regime_score = calculate_regime_score(quote_data)
        adjusted = apply_regime_dampening(ml_return, regime_score)
        if ml_return != 0:
            dampening = 1 - (adjusted / ml_return)
        else:
            dampening = 0
        if dampening > 0:
            logger.debug(
                f"{symbol}: ML={ml_return:.2%}, regime={regime_score:.2f}, "
                f"dampening={dampening:.1%}, adjusted={adjusted:.2%}"
            )
        return adjusted, regime_score, dampening

    security = await db.get_security(symbol)

    if not security or not security.get("quote_data"):
        logger.debug(f"{symbol}: No quote data, using raw ML prediction")
        return ml_return, 0.0, 0.0

    try:
        loaded_quote_data = json.loads(security["quote_data"])
    except (json.JSONDecodeError, TypeError):
        return ml_return, 0.0, 0.0

    regime_score = calculate_regime_score(loaded_quote_data)
    adjusted = apply_regime_dampening(ml_return, regime_score)

    if ml_return != 0:
        dampening = 1 - (adjusted / ml_return)
    else:
        dampening = 0

    if dampening > 0:
        logger.debug(
            f"{symbol}: ML={ml_return:.2%}, regime={regime_score:.2f}, "
            f"dampening={dampening:.1%}, adjusted={adjusted:.2%}"
        )

    return adjusted, regime_score, dampening
