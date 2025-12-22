"""Repository interface for stock data access."""

from abc import ABC, abstractmethod
from typing import Optional, List
from dataclasses import dataclass


@dataclass
class Stock:
    """Stock domain model."""
    symbol: str
    yahoo_symbol: Optional[str]
    name: str
    industry: Optional[str]
    geography: str
    priority_multiplier: float
    min_lot: int
    active: bool


class StockRepository(ABC):
    """Abstract repository for stock operations."""

    @abstractmethod
    async def get_by_symbol(self, symbol: str) -> Optional[Stock]:
        """Get stock by symbol."""
        pass

    @abstractmethod
    async def get_all_active(self) -> List[Stock]:
        """Get all active stocks."""
        pass

    @abstractmethod
    async def create(self, stock: Stock) -> None:
        """Create a new stock."""
        pass

    @abstractmethod
    async def update(self, symbol: str, **updates) -> None:
        """Update stock fields."""
        pass

    @abstractmethod
    async def delete(self, symbol: str) -> None:
        """Soft delete a stock (set active=False)."""
        pass

    @abstractmethod
    async def get_with_scores(self) -> List[dict]:
        """Get all active stocks with their scores and positions."""
        pass
