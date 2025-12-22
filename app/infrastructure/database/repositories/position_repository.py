"""SQLite implementation of PositionRepository."""

import aiosqlite
from typing import Optional, List
from datetime import datetime

from app.domain.repositories.position_repository import PositionRepository, Position


class SQLitePositionRepository(PositionRepository):
    """SQLite implementation of PositionRepository."""

    def __init__(self, db: aiosqlite.Connection):
        self.db = db

    async def get_by_symbol(self, symbol: str) -> Optional[Position]:
        """Get position by symbol."""
        cursor = await self.db.execute(
            "SELECT * FROM positions WHERE symbol = ?",
            (symbol.upper(),)
        )
        row = await cursor.fetchone()
        if not row:
            return None
        return Position(
            symbol=row["symbol"],
            quantity=row["quantity"],
            avg_price=row["avg_price"],
            current_price=row["current_price"],
            currency=row["currency"],
            currency_rate=row["currency_rate"],
            market_value_eur=row["market_value_eur"],
            last_updated=row["last_updated"],
        )

    async def get_all(self) -> List[Position]:
        """Get all positions."""
        cursor = await self.db.execute("SELECT * FROM positions")
        rows = await cursor.fetchall()
        return [
            Position(
                symbol=row["symbol"],
                quantity=row["quantity"],
                avg_price=row["avg_price"],
                current_price=row["current_price"],
                currency=row["currency"],
                currency_rate=row["currency_rate"],
                market_value_eur=row["market_value_eur"],
                last_updated=row["last_updated"],
            )
            for row in rows
        ]

    async def upsert(self, position: Position) -> None:
        """Insert or update a position."""
        await self.db.execute(
            """
            INSERT OR REPLACE INTO positions
            (symbol, quantity, avg_price, current_price, currency, currency_rate, market_value_eur, last_updated)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                position.symbol,
                position.quantity,
                position.avg_price,
                position.current_price,
                position.currency,
                position.currency_rate,
                position.market_value_eur,
                position.last_updated or datetime.now().isoformat(),
            ),
        )
        await self.db.commit()

    async def delete_all(self) -> None:
        """Delete all positions (used during sync)."""
        await self.db.execute("DELETE FROM positions")
        await self.db.commit()

    async def get_with_stock_info(self) -> List[dict]:
        """Get all positions with stock information."""
        cursor = await self.db.execute("""
            SELECT p.symbol, p.quantity, p.current_price, p.avg_price,
                   p.market_value_eur, s.name, s.geography, s.industry
            FROM positions p
            JOIN stocks s ON p.symbol = s.symbol
        """)
        rows = await cursor.fetchall()
        return [dict(row) for row in rows]
