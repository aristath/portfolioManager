"""SQLite implementation of StockRepository."""

import aiosqlite
from typing import Optional, List
from datetime import datetime

from app.domain.repositories.stock_repository import StockRepository, Stock
from app.domain.exceptions import StockNotFoundError


class SQLiteStockRepository(StockRepository):
    """SQLite implementation of StockRepository."""

    def __init__(self, db: aiosqlite.Connection):
        self.db = db

    async def get_by_symbol(self, symbol: str) -> Optional[Stock]:
        """Get stock by symbol."""
        cursor = await self.db.execute(
            "SELECT * FROM stocks WHERE symbol = ?",
            (symbol.upper(),)
        )
        row = await cursor.fetchone()
        if not row:
            return None
        return Stock(
            symbol=row["symbol"],
            yahoo_symbol=row["yahoo_symbol"],
            name=row["name"],
            industry=row["industry"],
            geography=row["geography"],
            priority_multiplier=row["priority_multiplier"] or 1.0,
            min_lot=row["min_lot"] or 1,
            active=bool(row["active"]),
        )

    async def get_all_active(self) -> List[Stock]:
        """Get all active stocks."""
        cursor = await self.db.execute(
            "SELECT * FROM stocks WHERE active = 1"
        )
        rows = await cursor.fetchall()
        return [
            Stock(
                symbol=row["symbol"],
                yahoo_symbol=row["yahoo_symbol"],
                name=row["name"],
                industry=row["industry"],
                geography=row["geography"],
                priority_multiplier=row["priority_multiplier"] or 1.0,
                min_lot=row["min_lot"] or 1,
                active=bool(row["active"]),
            )
            for row in rows
        ]

    async def create(self, stock: Stock) -> None:
        """Create a new stock."""
        await self.db.execute(
            """
            INSERT INTO stocks (symbol, yahoo_symbol, name, geography, industry, min_lot, active)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            """,
            (
                stock.symbol.upper(),
                stock.yahoo_symbol,
                stock.name,
                stock.geography.upper(),
                stock.industry,
                stock.min_lot,
                1 if stock.active else 0,
            ),
        )
        await self.db.commit()

    async def update(self, symbol: str, **updates) -> None:
        """Update stock fields."""
        if not updates:
            return

        updates_list = []
        values = []
        for key, value in updates.items():
            if key == "active":
                value = 1 if value else 0
            updates_list.append(f"{key} = ?")
            values.append(value)

        values.append(symbol.upper())
        await self.db.execute(
            f"UPDATE stocks SET {', '.join(updates_list)} WHERE symbol = ?",
            values
        )
        await self.db.commit()

    async def delete(self, symbol: str) -> None:
        """Soft delete a stock (set active=False)."""
        await self.db.execute(
            "UPDATE stocks SET active = 0 WHERE symbol = ?",
            (symbol.upper(),)
        )
        await self.db.commit()

    async def get_with_scores(self) -> List[dict]:
        """Get all active stocks with their scores and positions."""
        cursor = await self.db.execute("""
            SELECT s.*, sc.technical_score, sc.analyst_score,
                   sc.fundamental_score, sc.total_score, sc.volatility,
                   sc.calculated_at,
                   p.quantity as shares, p.current_price, p.avg_price,
                   p.market_value_eur as position_value
            FROM stocks s
            LEFT JOIN scores sc ON s.symbol = sc.symbol
            LEFT JOIN positions p ON s.symbol = p.symbol
            WHERE s.active = 1
        """)
        rows = await cursor.fetchall()
        return [dict(row) for row in rows]
