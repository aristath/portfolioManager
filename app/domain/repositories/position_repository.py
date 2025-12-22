"""Repository interface for position data access."""

from abc import ABC, abstractmethod
from typing import Optional, List
from dataclasses import dataclass


@dataclass
class Position:
    """Position domain model."""
    symbol: str
    quantity: float
    avg_price: float
    current_price: Optional[float]
    currency: str
    currency_rate: float
    market_value_eur: Optional[float]
    last_updated: Optional[str]


class PositionRepository(ABC):
    """Abstract repository for position operations."""

    @abstractmethod
    async def get_by_symbol(self, symbol: str) -> Optional[Position]:
        """Get position by symbol."""
        pass

    @abstractmethod
    async def get_all(self) -> List[Position]:
        """Get all positions."""
        pass

    @abstractmethod
    async def upsert(self, position: Position) -> None:
        """Insert or update a position."""
        pass

    @abstractmethod
    async def delete_all(self) -> None:
        """Delete all positions (used during sync)."""
        pass

    @abstractmethod
    async def get_with_stock_info(self) -> List[dict]:
        """Get all positions with stock information."""
        pass
