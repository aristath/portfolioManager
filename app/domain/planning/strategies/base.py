"""
Base Strategy Interface - Abstract base class for recommendation strategies.
"""

from abc import ABC, abstractmethod
from typing import List, Dict, Optional
from dataclasses import dataclass

from app.domain.scoring.models import PortfolioContext
from app.domain.models import Stock, Position


@dataclass
class StrategicGoal:
    """A strategic goal identified by a strategy."""
    strategy_type: str  # "diversification", "sustainability", "opportunity"
    category: str  # Strategy-specific (e.g., "geography", "quality", "value")
    name: str  # Specific identifier (e.g., "ASIA", "quality_score", "52w_high")
    action: str  # "increase" or "decrease"
    current_value: float  # Current metric value
    target_value: float  # Target metric value
    gap_size: float  # Absolute deviation
    priority_score: float  # Weighted by gap and impact
    target_value_change: float  # EUR amount needed to close gap
    description: str  # Human-readable goal description


class RecommendationStrategy(ABC):
    """Base interface for recommendation strategies."""
    
    @abstractmethod
    async def analyze_goals(
        self,
        portfolio_context: PortfolioContext,
        positions: List[Position],
        stocks: List[Stock],
        min_gap_threshold: float = 0.05
    ) -> List[StrategicGoal]:
        """
        Analyze portfolio and identify strategic goals.
        
        Args:
            portfolio_context: Current portfolio context
            positions: Current positions
            stocks: Available stocks
            min_gap_threshold: Minimum gap size to act on (default 5%)
        
        Returns:
            List of strategic goals, prioritized by priority_score
        """
        pass
    
    @abstractmethod
    async def find_best_buys(
        self,
        goals: List[StrategicGoal],
        portfolio_context: PortfolioContext,
        available_stocks: List[Stock],
        available_cash: float
    ) -> List[Dict]:
        """
        Find best stocks to buy for achieving goals.
        
        Args:
            goals: Strategic goals to achieve
            portfolio_context: Current portfolio context
            available_stocks: Stocks available to buy
            available_cash: Available cash in EUR
        
        Returns:
            List of dicts with keys: symbol, name, amount, quantity, price, reason
        """
        pass
    
    @abstractmethod
    async def find_best_sells(
        self,
        goals: List[StrategicGoal],
        portfolio_context: PortfolioContext,
        positions: List[Position],
        available_cash: float
    ) -> List[Dict]:
        """
        Find best positions to sell for achieving goals.
        
        Args:
            goals: Strategic goals to achieve
            portfolio_context: Current portfolio context
            positions: Current positions
            available_cash: Available cash in EUR
        
        Returns:
            List of dicts with keys: symbol, name, quantity, estimated_value, reason
        """
        pass
    
    @property
    @abstractmethod
    def strategy_name(self) -> str:
        """Strategy identifier (e.g., 'diversification', 'sustainability')."""
        pass
    
    @property
    @abstractmethod
    def strategy_description(self) -> str:
        """Human-readable description of strategy focus."""
        pass

