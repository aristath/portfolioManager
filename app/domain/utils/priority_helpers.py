"""Priority calculation helper functions.

These utilities are used for calculating stock priority scores based on
allocation weights, risk, and other factors.
"""


def calculate_weight_boost(weight: float) -> float:
    """
    Convert allocation weight (-1 to +1) to priority boost (0 to 1).

    Weight scale:
    - weight = +1 → boost = 1.0 (strong buy signal)
    - weight = 0 → boost = 0.5 (neutral)
    - weight = -1 → boost = 0.0 (avoid)

    Args:
        weight: Allocation weight from -1 to +1

    Returns:
        Priority boost from 0 to 1
    """
    # Clamp weight to valid range
    weight = max(-1, min(1, weight))
    # Linear mapping: -1 → 0, 0 → 0.5, +1 → 1.0
    return (weight + 1) / 2


def calculate_risk_adjustment(volatility: float) -> float:
    """
    Calculate risk adjustment factor based on volatility.

    Lower volatility = higher score.
    - 15% vol = 1.0 (best)
    - 50% vol = 0.0 (worst)
    - Unknown volatility = 0.5 (neutral)

    Args:
        volatility: Annualized volatility (0.0-1.0) or None

    Returns:
        Risk adjustment factor from 0 to 1
    """
    if volatility is None:
        return 0.5  # Neutral if unknown
    # 15% vol = 1.0, 50% vol = 0.0
    return max(0, min(1, 1 - (volatility - 0.15) / 0.35))
