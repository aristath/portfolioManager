"""Aggression calculator for dynamic position sizing.

Calculates how aggressively a satellite should trade based on:
1. Allocation status (percentage of target allocation achieved)
2. Drawdown status (distance from high water mark)

The aggression level (0.0-1.0) is used to scale position sizes:
- 1.0 = Full aggression (100% of strategy's normal position size)
- 0.8 = Reduced aggression (80% of normal)
- 0.6 = Conservative (60% of normal)
- 0.4 = Very conservative (40% of normal)
- 0.0 = Hibernation (no new trades)

The most conservative factor (min of allocation-based and drawdown-based) wins.
This prevents satellites from overtrading when underfunded or in drawdown.
"""

from dataclasses import dataclass
from typing import Optional


@dataclass
class AggressionResult:
    """Result of aggression calculation.

    Attributes:
        aggression: Final aggression level (0.0-1.0)
        allocation_aggression: Aggression based on allocation (0.0-1.0)
        drawdown_aggression: Aggression based on drawdown (0.0-1.0)
        limiting_factor: Which factor limited aggression ('allocation' or 'drawdown')
        current_value: Current bucket value
        target_value: Target bucket value
        pct_of_target: Current as % of target (0.0-1.0+)
        drawdown: Current drawdown from high water mark (0.0-1.0)
        in_hibernation: True if aggression is 0.0
    """

    aggression: float
    allocation_aggression: float
    drawdown_aggression: float
    limiting_factor: str
    current_value: float
    target_value: float
    pct_of_target: float
    drawdown: float
    in_hibernation: bool


def calculate_aggression(
    current_value: float,
    target_value: float,
    high_water_mark: Optional[float] = None,
) -> AggressionResult:
    """Calculate aggression level for a satellite bucket.

    Args:
        current_value: Current total value of bucket (positions + cash)
        target_value: Target allocation value for this bucket
        high_water_mark: Highest value achieved (for drawdown calculation)

    Returns:
        AggressionResult with final aggression and breakdown
    """
    # Percentage-based aggression (allocation status)
    pct_of_target = current_value / target_value if target_value > 0 else 0.0

    if pct_of_target >= 1.0:
        agg_pct = 1.0  # At or above target → full aggression
    elif pct_of_target >= 0.8:
        agg_pct = 0.8  # 80-100% of target → reduced aggression
    elif pct_of_target >= 0.6:
        agg_pct = 0.6  # 60-80% of target → conservative
    elif pct_of_target >= 0.4:
        agg_pct = 0.4  # 40-60% of target → very conservative
    else:
        agg_pct = 0.0  # Below 40% → hibernation

    # Drawdown-based aggression (risk management)
    if high_water_mark and high_water_mark > 0 and current_value < high_water_mark:
        drawdown = (high_water_mark - current_value) / high_water_mark
    else:
        drawdown = 0.0  # No drawdown (at or above high water mark)

    if drawdown >= 0.35:
        agg_dd = 0.0  # Severe drawdown (≥35%) → hibernation
    elif drawdown >= 0.25:
        agg_dd = 0.3  # Major drawdown (25-35%) → minimal trading
    elif drawdown >= 0.15:
        agg_dd = 0.7  # Moderate drawdown (15-25%) → reduced trading
    else:
        agg_dd = 1.0  # Minimal drawdown (<15%) → full aggression

    # Most conservative wins
    final_aggression = min(agg_pct, agg_dd)

    # Determine limiting factor
    if agg_pct < agg_dd:
        limiting_factor = "allocation"
    elif agg_dd < agg_pct:
        limiting_factor = "drawdown"
    else:
        limiting_factor = "equal"  # Both factors agree

    return AggressionResult(
        aggression=final_aggression,
        allocation_aggression=agg_pct,
        drawdown_aggression=agg_dd,
        limiting_factor=limiting_factor,
        current_value=current_value,
        target_value=target_value,
        pct_of_target=pct_of_target,
        drawdown=drawdown,
        in_hibernation=(final_aggression == 0.0),
    )


def should_hibernate(result: AggressionResult) -> bool:
    """Check if satellite should hibernate (aggression = 0.0).

    Args:
        result: AggressionResult from calculate_aggression

    Returns:
        True if satellite should not make any new trades
    """
    return result.in_hibernation


def scale_position_size(base_size: float, aggression: float) -> float:
    """Scale a position size by aggression level.

    Args:
        base_size: Base position size (from strategy parameters)
        aggression: Aggression level (0.0-1.0)

    Returns:
        Scaled position size
    """
    return base_size * aggression


def get_aggression_description(result: AggressionResult) -> str:
    """Get human-readable description of aggression status.

    Args:
        result: AggressionResult from calculate_aggression

    Returns:
        Description string
    """
    if result.in_hibernation:
        if result.limiting_factor == "allocation":
            return (
                f"HIBERNATION: Bucket at {result.pct_of_target*100:.1f}% of target "
                f"(below 40% threshold)"
            )
        elif result.limiting_factor == "drawdown":
            return (
                f"HIBERNATION: Drawdown at {result.drawdown*100:.1f}% "
                f"(above 35% threshold)"
            )
        else:
            return "HIBERNATION: Both allocation and drawdown in critical zone"

    status = []

    # Allocation status
    if result.pct_of_target >= 1.0:
        status.append(f"Fully funded ({result.pct_of_target*100:.1f}% of target)")
    else:
        status.append(f"Funding: {result.pct_of_target*100:.1f}% of target")

    # Drawdown status
    if result.drawdown > 0:
        status.append(f"Drawdown: {result.drawdown*100:.1f}%")
    else:
        status.append("No drawdown (at/above high water mark)")

    # Aggression level
    status.append(f"Aggression: {result.aggression*100:.0f}%")

    # Limiting factor
    if result.limiting_factor != "equal":
        status.append(f"(limited by {result.limiting_factor})")

    return " | ".join(status)
