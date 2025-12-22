"""Repository interface for stock score data access."""

from abc import ABC, abstractmethod
from typing import List, Optional
from dataclasses import dataclass
from datetime import datetime


@dataclass
class StockScore:
    """Stock score domain model."""
    symbol: str
    technical_score: Optional[float]
    analyst_score: Optional[float]
    fundamental_score: Optional[float]
    total_score: Optional[float]
    volatility: Optional[float]
    calculated_at: Optional[datetime]


class ScoreRepository(ABC):
    """Abstract repository for score operations."""

    @abstractmethod
    async def get_by_symbol(self, symbol: str) -> Optional[StockScore]:
        """Get score by symbol."""
        pass

    @abstractmethod
    async def upsert(self, score: StockScore) -> None:
        """Insert or update a score."""
        pass

    @abstractmethod
    async def get_all(self) -> List[StockScore]:
        """Get all scores."""
        pass
