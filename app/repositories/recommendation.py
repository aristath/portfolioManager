"""Recommendation repository - CRUD operations for recommendations table."""

import uuid
from datetime import datetime
from typing import List, Optional

from app.infrastructure.database import get_db_manager


class RecommendationRepository:
    """Repository for recommendation operations."""

    def __init__(self):
        self._db = get_db_manager().config

    async def find_existing(
        self, symbol: str, side: str, amount: float, reason: str
    ) -> Optional[dict]:
        """Find existing recommendation by matching criteria."""
        row = await self._db.fetchone(
            """
            SELECT * FROM recommendations
            WHERE symbol = ? AND side = ? AND amount = ? AND reason = ?
            """,
            (symbol.upper(), side.upper(), amount, reason),
        )
        return dict(row) if row else None

    async def create_or_update(self, recommendation_data: dict) -> Optional[str]:
        """
        Create new or update existing recommendation.

        Returns UUID if recommendation should be included, None if dismissed.
        """
        symbol = recommendation_data["symbol"].upper()
        side = recommendation_data["side"].upper()
        amount = recommendation_data["amount"]
        reason = recommendation_data["reason"]

        # Check if recommendation already exists
        existing = await self.find_existing(symbol, side, amount, reason)

        now = datetime.now().isoformat()

        if existing:
            existing_status = existing["status"]
            existing_uuid = existing["uuid"]

            # If dismissed, skip it (don't return UUID)
            if existing_status == "dismissed":
                return None

            # If pending, update timestamp and return existing UUID
            if existing_status == "pending":
                async with self._db.transaction() as conn:
                    await conn.execute(
                        """
                        UPDATE recommendations
                        SET updated_at = ?,
                            name = ?,
                            quantity = ?,
                            estimated_price = ?,
                            estimated_value = ?,
                            geography = ?,
                            industry = ?,
                            currency = ?,
                            priority = ?,
                            current_portfolio_score = ?,
                            new_portfolio_score = ?,
                            score_change = ?
                        WHERE uuid = ?
                        """,
                        (
                            now,
                            recommendation_data.get("name"),
                            recommendation_data.get("quantity"),
                            recommendation_data.get("estimated_price"),
                            recommendation_data.get("estimated_value"),
                            recommendation_data.get("geography"),
                            recommendation_data.get("industry"),
                            recommendation_data.get("currency", "EUR"),
                            recommendation_data.get("priority"),
                            recommendation_data.get("current_portfolio_score"),
                            recommendation_data.get("new_portfolio_score"),
                            recommendation_data.get("score_change"),
                            existing_uuid,
                        ),
                    )
                return existing_uuid

            # If executed, update back to pending (it's being regenerated, so it's valid again)
            # This handles the case where a recommendation was executed but then regenerated
            if existing_status == "executed":
                async with self._db.transaction() as conn:
                    await conn.execute(
                        """
                        UPDATE recommendations
                        SET status = 'pending',
                            updated_at = ?,
                            executed_at = NULL,
                            name = ?,
                            quantity = ?,
                            estimated_price = ?,
                            estimated_value = ?,
                            geography = ?,
                            industry = ?,
                            currency = ?,
                            priority = ?,
                            current_portfolio_score = ?,
                            new_portfolio_score = ?,
                            score_change = ?
                        WHERE uuid = ?
                        """,
                        (
                            now,
                            recommendation_data.get("name"),
                            recommendation_data.get("quantity"),
                            recommendation_data.get("estimated_price"),
                            recommendation_data.get("estimated_value"),
                            recommendation_data.get("geography"),
                            recommendation_data.get("industry"),
                            recommendation_data.get("currency", "EUR"),
                            recommendation_data.get("priority"),
                            recommendation_data.get("current_portfolio_score"),
                            recommendation_data.get("new_portfolio_score"),
                            recommendation_data.get("score_change"),
                            existing_uuid,
                        ),
                    )
                return existing_uuid

        # Create new recommendation
        new_uuid = str(uuid.uuid4())
        async with self._db.transaction() as conn:
            await conn.execute(
                """
                INSERT INTO recommendations
                (uuid, symbol, name, side, amount, quantity, estimated_price,
                 estimated_value, reason, geography, industry, currency, priority,
                 current_portfolio_score, new_portfolio_score, score_change,
                 status, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    new_uuid,
                    symbol,
                    recommendation_data.get("name"),
                    side,
                    amount,
                    recommendation_data.get("quantity"),
                    recommendation_data.get("estimated_price"),
                    recommendation_data.get("estimated_value"),
                    reason,
                    recommendation_data.get("geography"),
                    recommendation_data.get("industry"),
                    recommendation_data.get("currency", "EUR"),
                    recommendation_data.get("priority"),
                    recommendation_data.get("current_portfolio_score"),
                    recommendation_data.get("new_portfolio_score"),
                    recommendation_data.get("score_change"),
                    "pending",
                    now,
                    now,
                ),
            )
        return new_uuid

    async def get_by_uuid(self, uuid: str) -> Optional[dict]:
        """Get recommendation by UUID."""
        row = await self._db.fetchone(
            "SELECT * FROM recommendations WHERE uuid = ?", (uuid,)
        )
        return dict(row) if row else None

    async def get_pending(self, limit: int = 10) -> List[dict]:
        """Get pending recommendations (status='pending'), ordered by updated_at DESC."""
        rows = await self._db.fetchall(
            """
            SELECT * FROM recommendations
            WHERE status = 'pending'
            ORDER BY updated_at DESC
            LIMIT ?
            """,
            (limit,),
        )
        return [dict(row) for row in rows]

    async def get_pending_by_side(self, side: str, limit: int = 10) -> List[dict]:
        """Get pending recommendations by side (BUY or SELL)."""
        rows = await self._db.fetchall(
            """
            SELECT * FROM recommendations
            WHERE status = 'pending' AND side = ?
            ORDER BY updated_at DESC
            LIMIT ?
            """,
            (side.upper(), limit),
        )
        return [dict(row) for row in rows]

    async def mark_executed(self, uuid: str, executed_at: Optional[str] = None) -> None:
        """Mark recommendation as executed."""
        if executed_at is None:
            executed_at = datetime.now().isoformat()

        async with self._db.transaction() as conn:
            await conn.execute(
                """
                UPDATE recommendations
                SET status = 'executed', executed_at = ?
                WHERE uuid = ?
                """,
                (executed_at, uuid),
            )

    async def mark_dismissed(self, uuid: str, dismissed_at: Optional[str] = None) -> None:
        """Mark recommendation as dismissed."""
        if dismissed_at is None:
            dismissed_at = datetime.now().isoformat()

        async with self._db.transaction() as conn:
            await conn.execute(
                """
                UPDATE recommendations
                SET status = 'dismissed', dismissed_at = ?
                WHERE uuid = ?
                """,
                (dismissed_at, uuid),
            )

    async def is_dismissed(
        self, symbol: str, side: str, amount: float, reason: str
    ) -> bool:
        """Check if this exact recommendation was dismissed."""
        row = await self._db.fetchone(
            """
            SELECT status FROM recommendations
            WHERE symbol = ? AND side = ? AND amount = ? AND reason = ?
            """,
            (symbol.upper(), side.upper(), amount, reason),
        )
        return row and row["status"] == "dismissed"

    async def find_matching_for_execution(
        self, symbol: str, side: str, amount: float
    ) -> List[dict]:
        """
        Find pending recommendations that match execution criteria.

        Used to mark recommendations as executed after trade execution.
        Matches on symbol, side, and approximate amount (within 1% tolerance).
        """
        # Match on symbol, side, and amount within 1% tolerance
        rows = await self._db.fetchall(
            """
            SELECT * FROM recommendations
            WHERE status = 'pending'
              AND symbol = ?
              AND side = ?
              AND amount >= ? * 0.99
              AND amount <= ? * 1.01
            ORDER BY updated_at DESC
            """,
            (symbol.upper(), side.upper(), amount, amount),
        )
        return [dict(row) for row in rows]

