"""Sell scoring helper functions.

Pure functions for calculating individual sell score components.
These functions are used by the main sell scoring orchestrator.
"""

from datetime import datetime
from typing import Optional, Dict

from app.domain.scoring.constants import (
    DEFAULT_MIN_HOLD_DAYS,
    DEFAULT_MAX_LOSS_THRESHOLD,
    TARGET_RETURN_MIN,
    TARGET_RETURN_MAX,
    CONCENTRATION_HIGH,
    CONCENTRATION_MED,
    INSTABILITY_RATE_VERY_HOT,
    INSTABILITY_RATE_HOT,
    INSTABILITY_RATE_WARM,
    VOLATILITY_SPIKE_HIGH,
    VOLATILITY_SPIKE_MED,
    VOLATILITY_SPIKE_LOW,
    VALUATION_STRETCH_HIGH,
    VALUATION_STRETCH_MED,
    VALUATION_STRETCH_LOW,
)


def calculate_underperformance_score(
    current_price: float,
    avg_price: float,
    days_held: int,
    max_loss_threshold: float = DEFAULT_MAX_LOSS_THRESHOLD
) -> tuple:
    """
    Calculate underperformance score based on annualized return vs target.

    Returns:
        (score, profit_pct) tuple
    """
    if avg_price <= 0 or days_held <= 0:
        return 0.5, 0.0

    # Calculate profit percentage
    profit_pct = (current_price - avg_price) / avg_price

    # Calculate annualized return (CAGR)
    years_held = days_held / 365.0
    if years_held < 0.25:  # Less than 3 months - not enough data
        annualized_return = profit_pct  # Use simple return
    else:
        try:
            annualized_return = ((current_price / avg_price) ** (1 / years_held)) - 1
        except (ValueError, ZeroDivisionError):
            annualized_return = profit_pct

    # Score based on return vs target (8-15% annual ideal)
    # Higher score = more reason to sell
    if profit_pct < max_loss_threshold:
        # BLOCKED - loss too big
        return 0.0, profit_pct
    elif annualized_return < -0.05:
        # Loss of -5% to -20%: high sell priority (cut losses)
        return 0.9, profit_pct
    elif annualized_return < 0:
        # Small loss (-5% to 0%): stagnant, free up capital
        return 0.7, profit_pct
    elif annualized_return < TARGET_RETURN_MIN:
        # 0-8%: underperforming target
        return 0.5, profit_pct
    elif annualized_return <= TARGET_RETURN_MAX:
        # 8-15%: ideal range, don't sell
        return 0.1, profit_pct
    else:
        # >15%: exceeding target, consider taking profits
        return 0.3, profit_pct


def calculate_time_held_score(
    first_bought_at: Optional[str],
    min_hold_days: int = DEFAULT_MIN_HOLD_DAYS
) -> tuple:
    """
    Calculate time held score. Longer hold with underperformance = higher sell priority.

    Returns:
        (score, days_held) tuple
    """
    if not first_bought_at:
        # Unknown hold time - assume long enough
        return 0.6, 365

    try:
        bought_date = datetime.fromisoformat(first_bought_at.replace('Z', '+00:00'))
        if bought_date.tzinfo:
            bought_date = bought_date.replace(tzinfo=None)
        days_held = (datetime.now() - bought_date).days
    except (ValueError, TypeError):
        return 0.6, 365

    if days_held < min_hold_days:
        # BLOCKED - held less than 3 months
        return 0.0, days_held
    elif days_held < 180:
        # 3-6 months
        return 0.3, days_held
    elif days_held < 365:
        # 6-12 months
        return 0.6, days_held
    elif days_held < 730:
        # 12-24 months
        return 0.8, days_held
    else:
        # 24+ months - if still underperforming, time to cut
        return 1.0, days_held


def calculate_portfolio_balance_score(
    position_value: float,
    total_portfolio_value: float,
    geography: str,
    industry: str,
    geo_allocations: Dict[str, float],
    ind_allocations: Dict[str, float],
) -> float:
    """
    Calculate portfolio balance score based on overweight/underweight position.

    Higher score = position is overweight, more reason to trim.

    Returns:
        Score from 0.0 to 1.0
    """
    if total_portfolio_value <= 0:
        return 0.5

    position_pct = position_value / total_portfolio_value

    # Check geography concentration
    target_geo = geo_allocations.get(geography, 0.33)  # Default 33% if no target
    geo_excess = position_pct - target_geo if geography in geo_allocations else 0

    # Check industry concentration (use first industry if multiple)
    industries = [i.strip() for i in industry.split(',')] if industry else []
    target_ind = ind_allocations.get(industries[0], 0.10) if industries else 0.10
    ind_excess = position_pct - target_ind if industries and industries[0] in ind_allocations else 0

    # Use maximum excess (geography or industry)
    max_excess = max(geo_excess, ind_excess)

    # Score based on concentration
    if position_pct > CONCENTRATION_HIGH:  # >10% of portfolio
        return 1.0
    elif position_pct > CONCENTRATION_MED:  # >7% of portfolio
        return 0.7 + (position_pct - CONCENTRATION_MED) * 10  # 0.7 to 1.0
    elif max_excess > 0.03:  # >3% overweight
        return 0.5 + (max_excess / 0.05) * 0.2  # 0.5 to 0.7
    elif max_excess > 0:  # Slightly overweight
        return 0.3 + (max_excess / 0.03) * 0.2  # 0.3 to 0.5
    else:
        # Underweight or balanced - don't sell
        return 0.1


def calculate_instability_score(
    profit_pct: float,
    days_held: int,
    current_volatility: float,
    historical_volatility: float,
    distance_from_ma_200: float,
) -> float:
    """
    Detect potential instability/bubble conditions.
    High score = signs of unsustainable gains, consider trimming.

    Components:
    - Rate of gain (40%): Annualized return - penalize if unsustainably high
    - Volatility spike (30%): Current vs historical volatility
    - Valuation stretch (30%): Distance above 200-day MA
    """
    score = 0.0

    # 1. Rate of gain (40%)
    if days_held > 30:
        years = days_held / 365.0
        try:
            annualized = ((1 + profit_pct) ** (1 / years)) - 1 if years > 0 else profit_pct
        except (ValueError, OverflowError):
            annualized = profit_pct

        if annualized > INSTABILITY_RATE_VERY_HOT:  # >50% annualized = very hot
            rate_score = 1.0
        elif annualized > INSTABILITY_RATE_HOT:     # >30% annualized = hot
            rate_score = 0.7
        elif annualized > INSTABILITY_RATE_WARM:    # >20% annualized = warm
            rate_score = 0.4
        else:
            rate_score = 0.1  # Sustainable pace
    else:
        rate_score = 0.5  # Too early to tell
    score += rate_score * 0.40

    # 2. Volatility spike (30%)
    if historical_volatility > 0:
        vol_ratio = current_volatility / historical_volatility
        if vol_ratio > VOLATILITY_SPIKE_HIGH:     # Vol doubled
            vol_score = 1.0
        elif vol_ratio > VOLATILITY_SPIKE_MED:    # Vol up 50%
            vol_score = 0.7
        elif vol_ratio > VOLATILITY_SPIKE_LOW:    # Vol up 20%
            vol_score = 0.4
        else:
            vol_score = 0.1  # Normal volatility
    else:
        vol_score = 0.3  # No historical data - neutral
    score += vol_score * 0.30

    # 3. Valuation stretch (30%)
    if distance_from_ma_200 > VALUATION_STRETCH_HIGH:    # >30% above MA
        valuation_score = 1.0
    elif distance_from_ma_200 > VALUATION_STRETCH_MED:   # >20% above MA
        valuation_score = 0.7
    elif distance_from_ma_200 > VALUATION_STRETCH_LOW:   # >10% above MA
        valuation_score = 0.4
    else:
        valuation_score = 0.1  # Near or below MA
    score += valuation_score * 0.30

    # Floor for extreme profits (safety net)
    if profit_pct > 1.0:  # >100% gain
        score = max(score, 0.2)
    elif profit_pct > 0.75:  # >75% gain
        score = max(score, 0.1)

    return score

