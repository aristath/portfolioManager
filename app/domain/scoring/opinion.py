"""
Opinion Score - External analyst opinions and price targets.

Components:
- Recommendation (60%): Buy/Hold/Sell consensus
- Price Target (40%): Upside potential
"""

import logging
from typing import Optional

from app.services import yahoo

logger = logging.getLogger(__name__)


def calculate_opinion_score(
    symbol: str,
    yahoo_symbol: str = None
) -> float:
    """
    Calculate opinion score from analyst recommendations and price targets.

    Args:
        symbol: Tradernet symbol
        yahoo_symbol: Optional explicit Yahoo symbol override

    Returns:
        Score from 0 to 1.0
    """
    try:
        data = yahoo.get_analyst_data(symbol, yahoo_symbol=yahoo_symbol)

        if not data:
            return 0.5  # Neutral if no data

        # Recommendation score (already 0-1 from yahoo service)
        recommendation_score = data.recommendation_score

        # Target score: based on upside potential
        # 0% upside = 0.5, 20%+ upside = 1.0, -20% = 0.0
        upside = data.upside_pct / 100  # Convert to decimal
        target_score = 0.5 + (upside * 2.5)  # Scale
        target_score = max(0, min(1, target_score))

        # Combined (60% recommendation, 40% target)
        total = recommendation_score * 0.60 + target_score * 0.40

        return round(total, 3)

    except Exception as e:
        logger.error(f"Failed to calculate opinion score for {symbol}: {e}")
        return 0.5
