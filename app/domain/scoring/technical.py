"""
Technical Indicators - EMA, RSI, Bollinger, Sharpe, Max Drawdown.

Uses pandas-ta for technical indicators and empyrical for risk metrics.
"""

import logging
from typing import Optional, Tuple

import numpy as np
import pandas as pd
import pandas_ta as ta
import empyrical

from app.domain.scoring.constants import (
    EMA_LENGTH,
    RSI_LENGTH,
    BOLLINGER_LENGTH,
    BOLLINGER_STD,
    TRADING_DAYS_PER_YEAR,
)

logger = logging.getLogger(__name__)


def calculate_ema(closes: np.ndarray, length: int = EMA_LENGTH) -> Optional[float]:
    """
    Calculate Exponential Moving Average.

    Args:
        closes: Array of closing prices
        length: EMA period (default 200)

    Returns:
        Current EMA value or None if insufficient data
    """
    if len(closes) < length:
        # Fallback to SMA if not enough data for EMA
        return float(np.mean(closes))

    series = pd.Series(closes)
    ema = ta.ema(series, length=length)

    if ema is not None and len(ema) > 0 and not pd.isna(ema.iloc[-1]):
        return float(ema.iloc[-1])

    # Fallback to SMA
    return float(np.mean(closes[-length:]))


def calculate_rsi(closes: np.ndarray, length: int = RSI_LENGTH) -> Optional[float]:
    """
    Calculate Relative Strength Index.

    Args:
        closes: Array of closing prices
        length: RSI period (default 14)

    Returns:
        Current RSI value (0-100) or None if insufficient data
    """
    if len(closes) < length + 1:
        return None

    series = pd.Series(closes)
    rsi = ta.rsi(series, length=length)

    if rsi is not None and len(rsi) > 0 and not pd.isna(rsi.iloc[-1]):
        return float(rsi.iloc[-1])

    return None


def calculate_bollinger_bands(
    closes: np.ndarray,
    length: int = BOLLINGER_LENGTH,
    std: float = BOLLINGER_STD
) -> Optional[Tuple[float, float, float]]:
    """
    Calculate Bollinger Bands.

    Args:
        closes: Array of closing prices
        length: BB period (default 20)
        std: Standard deviation multiplier (default 2)

    Returns:
        Tuple of (lower, middle, upper) or None if insufficient data
    """
    if len(closes) < length:
        return None

    series = pd.Series(closes)
    bbands = ta.bbands(series, length=length, std=std)

    if bbands is None:
        return None

    # Dynamic column detection for version compatibility
    bb_lower_cols = [c for c in bbands.columns if c.startswith('BBL_')]
    bb_mid_cols = [c for c in bbands.columns if c.startswith('BBM_')]
    bb_upper_cols = [c for c in bbands.columns if c.startswith('BBU_')]

    if not (bb_lower_cols and bb_mid_cols and bb_upper_cols):
        return None

    lower = bbands[bb_lower_cols[0]].iloc[-1]
    middle = bbands[bb_mid_cols[0]].iloc[-1]
    upper = bbands[bb_upper_cols[0]].iloc[-1]

    if pd.isna(lower) or pd.isna(middle) or pd.isna(upper):
        return None

    return float(lower), float(middle), float(upper)


def calculate_bollinger_position(closes: np.ndarray) -> float:
    """
    Calculate position within Bollinger Bands.

    Returns:
        Position from 0 (at lower band) to 1 (at upper band).
        Returns 0.5 if bands can't be calculated.
    """
    bands = calculate_bollinger_bands(closes)
    if bands is None:
        return 0.5

    lower, _, upper = bands
    current = closes[-1]

    if upper <= lower:
        return 0.5

    position = (current - lower) / (upper - lower)
    return max(0.0, min(1.0, position))


def calculate_volatility(
    closes: np.ndarray,
    annualize: bool = True
) -> Optional[float]:
    """
    Calculate annualized volatility using empyrical.

    Args:
        closes: Array of closing prices
        annualize: Whether to annualize (default True)

    Returns:
        Annualized volatility or None if insufficient data
    """
    if len(closes) < 30:
        return None

    # Validate no zero/negative prices
    if np.any(closes[:-1] <= 0):
        logger.debug("Zero/negative prices detected, skipping volatility")
        return None

    returns = np.diff(closes) / closes[:-1]

    try:
        vol = float(empyrical.annual_volatility(returns))
        if not np.isfinite(vol) or vol < 0:
            return None
        return vol
    except Exception as e:
        logger.debug(f"Volatility calculation failed: {e}")
        return None


def calculate_sharpe_ratio(
    closes: np.ndarray,
    risk_free_rate: float = 0.0
) -> Optional[float]:
    """
    Calculate Sharpe ratio using empyrical.

    Args:
        closes: Array of closing prices
        risk_free_rate: Risk-free rate (default 0)

    Returns:
        Annualized Sharpe ratio or None if insufficient data
    """
    if len(closes) < 50:
        return None

    # Validate no zero/negative prices
    if np.any(closes[:-1] <= 0):
        logger.debug("Zero/negative prices detected, skipping Sharpe")
        return None

    returns = np.diff(closes) / closes[:-1]

    try:
        sharpe = float(empyrical.sharpe_ratio(
            returns,
            risk_free=risk_free_rate,
            annualization=TRADING_DAYS_PER_YEAR
        ))
        if not np.isfinite(sharpe):
            return None
        return sharpe
    except Exception as e:
        logger.debug(f"Sharpe ratio calculation failed: {e}")
        return None


def calculate_max_drawdown(closes: np.ndarray) -> Optional[float]:
    """
    Calculate maximum drawdown using empyrical.

    Args:
        closes: Array of closing prices

    Returns:
        Maximum drawdown as negative percentage (e.g., -0.20 = 20% drawdown)
        or None if insufficient data
    """
    if len(closes) < 50:
        return None

    # Validate no zero/negative prices
    if np.any(closes[:-1] <= 0):
        logger.debug("Zero/negative prices detected, skipping max drawdown")
        return None

    returns = np.diff(closes) / closes[:-1]

    try:
        mdd = float(empyrical.max_drawdown(returns))
        if not np.isfinite(mdd):
            return None
        return mdd  # Already negative
    except Exception as e:
        logger.debug(f"Max drawdown calculation failed: {e}")
        return None


def get_52_week_high(highs: np.ndarray) -> float:
    """
    Get 52-week high price.

    Args:
        highs: Array of high prices (at least 252 days for full year)

    Returns:
        52-week high price
    """
    if len(highs) >= 252:
        return float(max(highs[-252:]))
    return float(max(highs))


def get_52_week_low(lows: np.ndarray) -> float:
    """
    Get 52-week low price.

    Args:
        lows: Array of low prices (at least 252 days for full year)

    Returns:
        52-week low price
    """
    if len(lows) >= 252:
        return float(min(lows[-252:]))
    return float(min(lows))


def calculate_distance_from_ma(
    current_price: float,
    ma_value: float
) -> float:
    """
    Calculate percentage distance from moving average.

    Args:
        current_price: Current price
        ma_value: Moving average value

    Returns:
        Percentage distance (positive = above, negative = below)
    """
    if ma_value <= 0:
        return 0.0

    return (current_price - ma_value) / ma_value
