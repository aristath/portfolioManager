"""
Strategy Registry - Central registry for all recommendation strategies.
"""

from typing import Dict, Optional
from app.domain.planning.strategies.base import RecommendationStrategy
from app.domain.planning.strategies.diversification import DiversificationStrategy
from app.domain.planning.strategies.sustainability import SustainabilityStrategy
from app.domain.planning.strategies.opportunity import OpportunityStrategy

# Registry of all available strategies
_STRATEGIES: Dict[str, RecommendationStrategy] = {
    "diversification": DiversificationStrategy(),
    "sustainability": SustainabilityStrategy(),
    "opportunity": OpportunityStrategy(),
}


def get_strategy(name: str) -> Optional[RecommendationStrategy]:
    """Get strategy by name."""
    return _STRATEGIES.get(name)


def list_strategies() -> Dict[str, str]:
    """List all available strategies with descriptions."""
    return {
        name: strategy.strategy_description
        for name, strategy in _STRATEGIES.items()
    }


__all__ = ["get_strategy", "list_strategies", "RecommendationStrategy"]

