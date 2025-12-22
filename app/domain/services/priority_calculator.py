"""Priority calculation service.

Pure business logic for calculating stock priority scores based on
quality, allocation weights, diversification, and risk.
"""

from typing import Dict, List, Optional
from dataclasses import dataclass

from app.domain.utils.priority_helpers import (
    calculate_weight_boost,
    calculate_risk_adjustment,
)


@dataclass
class PriorityInput:
    """Input data for priority calculation."""
    symbol: str
    name: str
    geography: str
    industry: Optional[str]
    stock_score: float
    volatility: Optional[float]
    multiplier: float
    position_value: float
    total_portfolio_value: float


@dataclass
class PriorityResult:
    """Result of priority calculation."""
    symbol: str
    name: str
    geography: str
    industry: str
    stock_score: float
    volatility: Optional[float]
    multiplier: float
    geo_need: float  # How prioritized is this geography (0 to 1)
    industry_need: float  # How prioritized is this industry (0 to 1)
    combined_priority: float  # Final priority score


class PriorityCalculator:
    """Service for calculating stock priority scores."""

    @staticmethod
    def parse_industries(industry_str: Optional[str]) -> List[str]:
        """
        Parse comma-separated industry string into list.

        Args:
            industry_str: Comma-separated industries (e.g., "Industrial, Defense")

        Returns:
            List of industry names, or empty list if None/empty
        """
        if not industry_str:
            return []
        return [ind.strip() for ind in industry_str.split(",") if ind.strip()]

    @staticmethod
    def calculate_priority(
        input_data: PriorityInput,
        geo_weights: Dict[str, float],
        industry_weights: Dict[str, float],
    ) -> PriorityResult:
        """
        Calculate priority score for a stock.

        Strategy:
        1. Quality (40%): stock score + conviction boost
        2. Allocation Weight (30%): boost from geo/industry weights
        3. Diversification (15%): penalty for concentrated positions
        4. Risk (15%): volatility-based adjustment
        5. Apply manual multiplier

        Args:
            input_data: Stock data for priority calculation
            geo_weights: Geography weights (name -> weight from -1 to +1)
            industry_weights: Industry weights (name -> weight from -1 to +1)

        Returns:
            PriorityResult with calculated priority
        """
        industries = PriorityCalculator.parse_industries(input_data.industry)

        # 1. Quality with conviction boost (high scorers get extra)
        conviction_boost = max(0, (input_data.stock_score - 0.5) * 0.4) if input_data.stock_score > 0.5 else 0
        quality = input_data.stock_score + conviction_boost

        # 2. Allocation weight boost
        # Get weight for this stock's geography and industries
        # Weight ranges from -1 (avoid) to +1 (prioritize), 0 = neutral
        geo_weight = geo_weights.get(input_data.geography, 0)  # Default 0 = neutral
        geo_boost = calculate_weight_boost(geo_weight)
        geo_need = geo_boost  # For logging (higher = more desired)

        ind_boost = 0.5  # Default neutral
        industry_need = 0.5
        if industries:
            ind_weights = [industry_weights.get(ind, 0) for ind in industries]
            ind_boosts = [calculate_weight_boost(w) for w in ind_weights]
            ind_boost = sum(ind_boosts) / len(ind_boosts)
            industry_need = ind_boost

        # Combined allocation boost (weighted average of geo and industry)
        allocation_boost = geo_boost * 0.6 + ind_boost * 0.4

        # 3. Diversification penalty (reduce priority for concentrated positions)
        position_pct = input_data.position_value / input_data.total_portfolio_value if input_data.total_portfolio_value > 0 else 0
        # Higher weight = ok to have more, lower weight = penalize concentration more
        geo_concentration_penalty = position_pct * (1 - geo_boost)  # More penalty if weight is low
        diversification = 1.0 - min(0.5, geo_concentration_penalty * 3)

        # 4. Risk adjustment based on volatility
        risk_adj = calculate_risk_adjustment(input_data.volatility)

        # Weighted combination
        raw_priority = (
            quality * 0.40 +
            allocation_boost * 0.30 +
            diversification * 0.15 +
            risk_adj * 0.15
        )

        # Apply manual multiplier
        combined_priority = raw_priority * input_data.multiplier

        return PriorityResult(
            symbol=input_data.symbol,
            name=input_data.name,
            geography=input_data.geography,
            industry=input_data.industry or "Unknown",
            stock_score=input_data.stock_score,
            volatility=input_data.volatility,
            multiplier=input_data.multiplier,
            geo_need=round(geo_need, 4),
            industry_need=round(industry_need, 4),
            combined_priority=round(combined_priority, 4),
        )

    @staticmethod
    def calculate_priorities(
        inputs: List[PriorityInput],
        geo_weights: Dict[str, float],
        industry_weights: Dict[str, float],
    ) -> List[PriorityResult]:
        """
        Calculate priorities for multiple stocks.

        Args:
            inputs: List of stock data for priority calculation
            geo_weights: Geography weights (name -> weight from -1 to +1)
            industry_weights: Industry weights (name -> weight from -1 to +1)

        Returns:
            List of PriorityResult sorted by combined_priority (highest first)
        """
        results = [
            PriorityCalculator.calculate_priority(input_data, geo_weights, industry_weights)
            for input_data in inputs
        ]

        # Sort by combined priority (highest first)
        results.sort(key=lambda x: x.combined_priority, reverse=True)

        return results
