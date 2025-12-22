"""Repository interface for trade data access."""

from abc import ABC, abstractmethod
from typing import List, Optional
from dataclasses import dataclass
from datetime import datetime


@dataclass
class Trade:
    """Trade domain model."""
    symbol: str
    side: str  # 'BUY' or 'SELL'
    quantity: float
    price: float
    executed_at: datetime
    order_id: Optional[str]


class TradeRepository(ABC):
    """Abstract repository for trade operations."""

    @abstractmethod
    async def create(self, trade: Trade) -> None:
        """Create a new trade record."""
        pass

    @abstractmethod
    async def get_history(self, limit: int = 50) -> List[Trade]:
        """Get trade history."""
        pass
