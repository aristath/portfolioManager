"""SQLite implementation of TradeRepository."""

import aiosqlite
from typing import List, Optional
from datetime import datetime

from app.domain.repositories.trade_repository import TradeRepository, Trade


class SQLiteTradeRepository(TradeRepository):
    """SQLite implementation of TradeRepository."""

    def __init__(self, db: aiosqlite.Connection):
        self.db = db

    async def create(self, trade: Trade) -> None:
        """Create a new trade record."""
        executed_at_str = trade.executed_at.isoformat() if isinstance(trade.executed_at, datetime) else str(trade.executed_at)

        await self.db.execute(
            """
            INSERT INTO trades (symbol, side, quantity, price, executed_at, order_id)
            VALUES (?, ?, ?, ?, ?, ?)
            """,
            (
                trade.symbol,
                trade.side,
                trade.quantity,
                trade.price,
                executed_at_str,
                trade.order_id,
            ),
        )
        await self.db.commit()

    async def get_history(self, limit: int = 50) -> List[Trade]:
        """Get trade history."""
        cursor = await self.db.execute("""
            SELECT t.*, s.name
            FROM trades t
            JOIN stocks s ON t.symbol = s.symbol
            ORDER BY t.executed_at DESC
            LIMIT ?
        """, (limit,))
        rows = await cursor.fetchall()
        trades = []
        for row in rows:
            executed_at = None
            if row["executed_at"]:
                try:
                    executed_at = datetime.fromisoformat(row["executed_at"])
                except (ValueError, TypeError):
                    executed_at = datetime.now()

            trades.append(Trade(
                symbol=row["symbol"],
                side=row["side"],
                quantity=row["quantity"],
                price=row["price"],
                executed_at=executed_at,
                order_id=row["order_id"],
            ))
        return trades
