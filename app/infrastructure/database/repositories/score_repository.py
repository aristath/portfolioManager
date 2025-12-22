"""SQLite implementation of ScoreRepository."""

import aiosqlite
from typing import Optional, List
from datetime import datetime

from app.domain.repositories.score_repository import ScoreRepository, StockScore


class SQLiteScoreRepository(ScoreRepository):
    """SQLite implementation of ScoreRepository."""

    def __init__(self, db: aiosqlite.Connection):
        self.db = db

    async def get_by_symbol(self, symbol: str) -> Optional[StockScore]:
        """Get score by symbol."""
        cursor = await self.db.execute(
            "SELECT * FROM scores WHERE symbol = ?",
            (symbol.upper(),)
        )
        row = await cursor.fetchone()
        if not row:
            return None

        calculated_at = None
        if row["calculated_at"]:
            try:
                calculated_at = datetime.fromisoformat(row["calculated_at"])
            except (ValueError, TypeError):
                pass

        return StockScore(
            symbol=row["symbol"],
            technical_score=row["technical_score"],
            analyst_score=row["analyst_score"],
            fundamental_score=row["fundamental_score"],
            total_score=row["total_score"],
            volatility=row["volatility"],
            calculated_at=calculated_at,
        )

    async def upsert(self, score: StockScore) -> None:
        """Insert or update a score."""
        calculated_at_str = None
        if score.calculated_at:
            if isinstance(score.calculated_at, datetime):
                calculated_at_str = score.calculated_at.isoformat()
            else:
                calculated_at_str = str(score.calculated_at)

        await self.db.execute(
            """
            INSERT OR REPLACE INTO scores
            (symbol, technical_score, analyst_score, fundamental_score,
             total_score, volatility, calculated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            """,
            (
                score.symbol,
                score.technical_score,
                score.analyst_score,
                score.fundamental_score,
                score.total_score,
                score.volatility,
                calculated_at_str,
            ),
        )
        await self.db.commit()

    async def get_all(self) -> List[StockScore]:
        """Get all scores."""
        cursor = await self.db.execute("SELECT * FROM scores")
        rows = await cursor.fetchall()
        scores = []
        for row in rows:
            calculated_at = None
            if row["calculated_at"]:
                try:
                    calculated_at = datetime.fromisoformat(row["calculated_at"])
                except (ValueError, TypeError):
                    pass

            scores.append(StockScore(
                symbol=row["symbol"],
                technical_score=row["technical_score"],
                analyst_score=row["analyst_score"],
                fundamental_score=row["fundamental_score"],
                total_score=row["total_score"],
                volatility=row["volatility"],
                calculated_at=calculated_at,
            ))
        return scores
