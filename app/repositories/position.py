"""Position repository - CRUD operations for positions table."""

from datetime import datetime
from typing import List, Optional

from app.domain.models import Position
from app.infrastructure.database import get_db_manager


class PositionRepository:
    """Repository for current position operations."""

    def __init__(self):
        self._db = get_db_manager().state

    async def get_by_symbol(self, symbol: str) -> Optional[Position]:
        """Get position by symbol."""
        row = await self._db.fetchone(
            "SELECT * FROM positions WHERE symbol = ?",
            (symbol.upper(),)
        )
        if not row:
            return None
        return self._row_to_position(row)

    async def get_all(self) -> List[Position]:
        """Get all positions."""
        rows = await self._db.fetchall("SELECT * FROM positions")
        return [self._row_to_position(row) for row in rows]

    async def upsert(self, position: Position) -> None:
        """Insert or update a position."""
        now = datetime.now().isoformat()

        async with self._db.transaction() as conn:
            await conn.execute(
                """
                INSERT OR REPLACE INTO positions
                (symbol, quantity, avg_price, current_price, currency,
                 currency_rate, market_value_eur, cost_basis_eur,
                 unrealized_pnl, unrealized_pnl_pct, last_updated,
                 first_bought_at, last_sold_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    position.symbol.upper(),
                    position.quantity,
                    position.avg_price,
                    position.current_price,
                    position.currency,
                    position.currency_rate,
                    position.market_value_eur,
                    position.cost_basis_eur,
                    position.unrealized_pnl,
                    position.unrealized_pnl_pct,
                    position.last_updated or now,
                    position.first_bought_at,
                    position.last_sold_at,
                )
            )

    async def delete_all(self) -> None:
        """Delete all positions (used during sync)."""
        async with self._db.transaction() as conn:
            await conn.execute("DELETE FROM positions")

    async def delete(self, symbol: str) -> None:
        """Delete a specific position."""
        async with self._db.transaction() as conn:
            await conn.execute(
                "DELETE FROM positions WHERE symbol = ?",
                (symbol.upper(),)
            )

    async def update_price(self, symbol: str, price: float, currency_rate: float = 1.0) -> None:
        """Update current price and recalculate market value."""
        now = datetime.now().isoformat()

        async with self._db.transaction() as conn:
            await conn.execute(
                """
                UPDATE positions SET
                    current_price = ?,
                    market_value_eur = quantity * ? / ?,
                    unrealized_pnl = (? - avg_price) * quantity / ?,
                    unrealized_pnl_pct = CASE
                        WHEN avg_price > 0 THEN ((? / avg_price) - 1) * 100
                        ELSE 0
                    END,
                    last_updated = ?
                WHERE symbol = ?
                """,
                (price, price, currency_rate, price, currency_rate, price, now, symbol.upper())
            )

    async def update_last_sold_at(self, symbol: str) -> None:
        """Update the last_sold_at timestamp after a sell."""
        now = datetime.now().isoformat()

        async with self._db.transaction() as conn:
            await conn.execute(
                "UPDATE positions SET last_sold_at = ? WHERE symbol = ?",
                (now, symbol.upper())
            )

    async def get_total_value(self) -> float:
        """Get total portfolio value in EUR."""
        row = await self._db.fetchone(
            "SELECT COALESCE(SUM(market_value_eur), 0) as total FROM positions"
        )
        return row["total"] if row else 0.0

    def _row_to_position(self, row) -> Position:
        """Convert database row to Position model."""
        return Position(
            symbol=row["symbol"],
            quantity=row["quantity"],
            avg_price=row["avg_price"],
            current_price=row["current_price"],
            currency=row["currency"] or "EUR",
            currency_rate=row["currency_rate"] or 1.0,
            market_value_eur=row["market_value_eur"],
            cost_basis_eur=row["cost_basis_eur"] if "cost_basis_eur" in row.keys() else None,
            unrealized_pnl=row["unrealized_pnl"] if "unrealized_pnl" in row.keys() else None,
            unrealized_pnl_pct=row["unrealized_pnl_pct"] if "unrealized_pnl_pct" in row.keys() else None,
            last_updated=row["last_updated"],
            first_bought_at=row["first_bought_at"] if "first_bought_at" in row.keys() else None,
            last_sold_at=row["last_sold_at"] if "last_sold_at" in row.keys() else None,
        )
