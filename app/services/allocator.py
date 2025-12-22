"""Portfolio allocation and rebalancing logic."""

import logging
from dataclasses import dataclass
from typing import Optional

# Removed aiosqlite import - no longer needed

from app.config import settings
from app.domain.utils.priority_helpers import (
    calculate_weight_boost,
    calculate_risk_adjustment,
)

logger = logging.getLogger(__name__)


def parse_industries(industry_str: str) -> list[str]:
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


@dataclass
class AllocationStatus:
    """Current allocation vs target."""
    category: str  # geography or industry
    name: str  # EU, ASIA, US or Technology, etc.
    target_pct: float
    current_pct: float
    current_value: float
    deviation: float  # current - target (negative = underweight)


@dataclass
class PortfolioSummary:
    """Complete portfolio allocation summary."""
    total_value: float
    cash_balance: float
    geographic_allocations: list[AllocationStatus]
    industry_allocations: list[AllocationStatus]


@dataclass
class TradeRecommendation:
    """Recommended trade for rebalancing."""
    symbol: str
    name: str
    side: str  # BUY or SELL
    quantity: float
    estimated_price: float
    estimated_value: float
    reason: str  # Why this trade is recommended


# Removed get_portfolio_summary() - use PortfolioService.get_portfolio_summary() instead

@dataclass
class StockPriority:
    """Priority score for a stock candidate."""
    symbol: str
    name: str
    geography: str
    industry: str
    stock_score: float
    volatility: float  # Raw volatility (0.0-1.0)
    multiplier: float  # Manual priority multiplier
    min_lot: int  # Minimum lot size for trading
    geo_need: float  # How underweight is this geography (0 to 1)
    industry_need: float  # How underweight is this industry (0 to 1)
    combined_priority: float  # Enhanced priority score


def calculate_diversification_penalty(
    position_pct: float,
    geo_overweight: float,
    industry_overweight: float
) -> float:
    """Penalty for concentrated positions."""
    position_penalty = min(0.3, position_pct * 3)  # 10% position = 0.3 penalty
    geo_penalty = max(0, geo_overweight * 0.5)
    industry_penalty = max(0, industry_overweight * 0.5)

    total_penalty = position_penalty * 0.4 + geo_penalty * 0.3 + industry_penalty * 0.3
    return min(0.5, total_penalty)  # Cap at 0.5


def calculate_position_size(
    candidate: StockPriority,
    base_size: float,
    min_size: float,
) -> float:
    """
    Calculate position size based on conviction and risk.

    Args:
        candidate: Stock priority data
        base_size: Base investment amount per trade
        min_size: Minimum trade size

    Returns:
        Adjusted position size (0.8x to 1.2x of base)
    """
    # Conviction multiplier: 0.8 to 1.2 based on stock score
    conviction_mult = 0.8 + (candidate.stock_score - 0.5) * 0.8
    conviction_mult = max(0.8, min(1.2, conviction_mult))

    # Priority multiplier: 0.9 to 1.1 based on combined priority
    priority_mult = 0.9 + (candidate.combined_priority / 3.0) * 0.2
    priority_mult = max(0.9, min(1.1, priority_mult))

    # Volatility penalty (if available)
    if candidate.volatility is not None:
        vol_mult = max(0.7, 1.0 - (candidate.volatility - 0.15) * 0.5)
    else:
        vol_mult = 1.0

    size = base_size * conviction_mult * priority_mult * vol_mult
    return max(min_size, min(size, base_size * 1.2))


def get_max_trades(cash: float) -> int:
    """Calculate maximum trades based on available cash."""
    return min(
        settings.max_trades_per_cycle,
        int(cash / settings.min_trade_size)
    )


# Removed calculate_rebalance_trades() - use RebalancingService.calculate_rebalance_trades() instead
# Removed execute_trades() - use TradeExecutionService.execute_trades() instead
