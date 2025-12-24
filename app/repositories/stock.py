"""Stock repository - CRUD operations for stocks table."""

from datetime import datetime
from typing import List, Optional

from app.domain.models import Stock
from app.infrastructure.database import get_db_manager


class StockRepository:
    """Repository for stock universe operations."""

    def __init__(self):
        self._db = get_db_manager().config

    async def get_by_symbol(self, symbol: str) -> Optional[Stock]:
        """Get stock by symbol."""
        row = await self._db.fetchone(
            "SELECT * FROM stocks WHERE symbol = ?",
            (symbol.upper(),)
        )
        if not row:
            return None
        return self._row_to_stock(row)

    async def get_all_active(self) -> List[Stock]:
        """Get all active stocks."""
        rows = await self._db.fetchall(
            "SELECT * FROM stocks WHERE active = 1"
        )
        return [self._row_to_stock(row) for row in rows]

    async def get_all(self) -> List[Stock]:
        """Get all stocks (active and inactive)."""
        rows = await self._db.fetchall("SELECT * FROM stocks")
        return [self._row_to_stock(row) for row in rows]

    async def create(self, stock: Stock) -> None:
        """Create a new stock."""
        now = datetime.now().isoformat()
        async with self._db.transaction() as conn:
            await conn.execute(
                """
                INSERT INTO stocks
                (symbol, yahoo_symbol, name, industry, geography,
                 priority_multiplier, min_lot, active, allow_buy, allow_sell,
                 currency, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    stock.symbol.upper(),
                    stock.yahoo_symbol,
                    stock.name,
                    stock.industry,
                    stock.geography.upper(),
                    stock.priority_multiplier,
                    stock.min_lot,
                    1 if stock.active else 0,
                    1 if stock.allow_buy else 0,
                    1 if stock.allow_sell else 0,
                    stock.currency,
                    now,
                    now,
                )
            )

    async def update(self, symbol: str, **updates) -> None:
        """Update stock fields."""
        if not updates:
            return

        now = datetime.now().isoformat()
        updates["updated_at"] = now

        # Convert booleans to integers
        for key in ("active", "allow_buy", "allow_sell"):
            if key in updates:
                updates[key] = 1 if updates[key] else 0

        set_clause = ", ".join(f"{k} = ?" for k in updates.keys())
        values = list(updates.values()) + [symbol.upper()]

        async with self._db.transaction() as conn:
            await conn.execute(
                f"UPDATE stocks SET {set_clause} WHERE symbol = ?",
                values
            )

    async def delete(self, symbol: str) -> None:
        """Soft delete a stock (set active=False)."""
        await self.update(symbol, active=False)

    async def get_with_scores(self) -> List[dict]:
        """Get all active stocks with their scores and positions."""
        # This joins across databases - needs to be handled differently
        # For now, return just stocks; scoring will join separately
        rows = await self._db.fetchall(
            "SELECT * FROM stocks WHERE active = 1"
        )
        return [dict(row) for row in rows]

    def _row_to_stock(self, row) -> Stock:
        """Convert database row to Stock model."""
        return Stock(
            symbol=row["symbol"],
            yahoo_symbol=row["yahoo_symbol"],
            name=row["name"],
            industry=row["industry"],
            geography=row["geography"],
            priority_multiplier=row["priority_multiplier"] or 1.0,
            min_lot=row["min_lot"] or 1,
            active=bool(row["active"]),
            allow_buy=bool(row["allow_buy"]) if row["allow_buy"] is not None else True,
            allow_sell=bool(row["allow_sell"]) if row["allow_sell"] is not None else False,
            currency=row["currency"],
        )
