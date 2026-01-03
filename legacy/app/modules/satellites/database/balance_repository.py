"""Balance repository - operations for bucket_balances and bucket_transactions tables."""

from datetime import datetime
from typing import Dict, List, Optional

from app.core.database.manager import get_db_manager
from app.modules.satellites.domain.enums import TransactionType
from app.modules.satellites.domain.models import BucketBalance, BucketTransaction


class BalanceRepository:
    """Repository for virtual cash balance operations.

    Tracks per-bucket, per-currency cash balances and
    maintains an audit trail of all cash flow transactions.
    """

    def __init__(self):
        self._db = get_db_manager().satellites

    # Balance operations

    async def get_balance(
        self, bucket_id: str, currency: str
    ) -> Optional[BucketBalance]:
        """Get balance for a specific bucket and currency."""
        row = await self._db.fetchone(
            """SELECT * FROM bucket_balances
               WHERE bucket_id = ? AND currency = ?""",
            (bucket_id, currency.upper()),
        )
        if not row:
            return None
        return self._row_to_balance(row)

    async def get_all_balances(self, bucket_id: str) -> List[BucketBalance]:
        """Get all currency balances for a bucket."""
        rows = await self._db.fetchall(
            "SELECT * FROM bucket_balances WHERE bucket_id = ? ORDER BY currency",
            (bucket_id,),
        )
        return [self._row_to_balance(row) for row in rows]

    async def get_all_balances_by_currency(self, currency: str) -> List[BucketBalance]:
        """Get balances for all buckets in a specific currency."""
        rows = await self._db.fetchall(
            "SELECT * FROM bucket_balances WHERE currency = ? ORDER BY bucket_id",
            (currency.upper(),),
        )
        return [self._row_to_balance(row) for row in rows]

    async def get_total_by_currency(self, currency: str) -> float:
        """Get sum of all bucket balances for a currency.

        This should equal the actual brokerage balance for that currency.
        """
        row = await self._db.fetchone(
            """SELECT COALESCE(SUM(balance), 0) as total
               FROM bucket_balances
               WHERE currency = ?""",
            (currency.upper(),),
        )
        return row["total"] if row else 0.0

    async def get_balance_amount(self, bucket_id: str, currency: str) -> float:
        """Get balance amount, returning 0 if not found."""
        balance = await self.get_balance(bucket_id, currency)
        return balance.balance if balance else 0.0

    async def set_balance(
        self, bucket_id: str, currency: str, amount: float
    ) -> BucketBalance:
        """Set balance to a specific amount (upsert)."""
        now = datetime.now().isoformat()
        currency = currency.upper()

        await self._db.execute(
            """INSERT OR REPLACE INTO bucket_balances
               (bucket_id, currency, balance, last_updated)
               VALUES (?, ?, ?, ?)""",
            (bucket_id, currency, amount, now),
        )
        await self._db.commit()

        return BucketBalance(
            bucket_id=bucket_id,
            currency=currency,
            balance=amount,
            last_updated=now,
        )

    async def adjust_balance(
        self, bucket_id: str, currency: str, delta: float
    ) -> BucketBalance:
        """Adjust balance by a delta amount.

        Creates the balance record if it doesn't exist.

        Args:
            bucket_id: Bucket ID
            currency: Currency code
            delta: Amount to add (positive) or subtract (negative)

        Returns:
            Updated balance
        """
        now = datetime.now().isoformat()
        currency = currency.upper()

        # First, ensure the row exists
        await self._db.execute(
            """INSERT OR IGNORE INTO bucket_balances
               (bucket_id, currency, balance, last_updated)
               VALUES (?, ?, 0, ?)""",
            (bucket_id, currency, now),
        )

        # Then update with delta
        await self._db.execute(
            """UPDATE bucket_balances
               SET balance = balance + ?,
                   last_updated = ?
               WHERE bucket_id = ? AND currency = ?""",
            (delta, now, bucket_id, currency),
        )
        await self._db.commit()

        balance = await self.get_balance(bucket_id, currency)
        if not balance:
            # Should not happen, but handle gracefully
            return BucketBalance(
                bucket_id=bucket_id,
                currency=currency,
                balance=delta,
                last_updated=now,
            )
        return balance

    async def delete_balances(self, bucket_id: str) -> int:
        """Delete all balances for a bucket.

        Returns:
            Number of balance records deleted
        """
        result = await self._db.execute(
            "DELETE FROM bucket_balances WHERE bucket_id = ?",
            (bucket_id,),
        )
        await self._db.commit()
        return result.rowcount

    # Transaction operations

    async def record_transaction(
        self, transaction: BucketTransaction
    ) -> BucketTransaction:
        """Record a transaction in the audit trail."""
        now = datetime.now().isoformat()
        transaction.created_at = now

        async with self._db.transaction() as conn:
            cursor = await conn.execute(
                """INSERT INTO bucket_transactions
                   (bucket_id, type, amount, currency, description, created_at)
                   VALUES (?, ?, ?, ?, ?, ?)""",
                (
                    transaction.bucket_id,
                    transaction.type.value,
                    transaction.amount,
                    transaction.currency.upper(),
                    transaction.description,
                    transaction.created_at,
                ),
            )
            transaction.id = cursor.lastrowid

        return transaction

    async def get_transactions(
        self,
        bucket_id: str,
        limit: int = 100,
        offset: int = 0,
        transaction_type: Optional[TransactionType] = None,
    ) -> List[BucketTransaction]:
        """Get transactions for a bucket with optional type filter."""
        if transaction_type:
            rows = await self._db.fetchall(
                """SELECT * FROM bucket_transactions
                   WHERE bucket_id = ? AND type = ?
                   ORDER BY created_at DESC
                   LIMIT ? OFFSET ?""",
                (bucket_id, transaction_type.value, limit, offset),
            )
        else:
            rows = await self._db.fetchall(
                """SELECT * FROM bucket_transactions
                   WHERE bucket_id = ?
                   ORDER BY created_at DESC
                   LIMIT ? OFFSET ?""",
                (bucket_id, limit, offset),
            )
        return [self._row_to_transaction(row) for row in rows]

    async def get_recent_transactions(
        self, bucket_id: str, days: int = 30
    ) -> List[BucketTransaction]:
        """Get transactions from the last N days."""
        rows = await self._db.fetchall(
            """SELECT * FROM bucket_transactions
               WHERE bucket_id = ?
                 AND created_at >= datetime('now', ? || ' days')
               ORDER BY created_at DESC""",
            (bucket_id, -days),
        )
        return [self._row_to_transaction(row) for row in rows]

    async def get_transactions_by_type(
        self, transaction_type: TransactionType, limit: int = 100
    ) -> List[BucketTransaction]:
        """Get all transactions of a specific type across all buckets."""
        rows = await self._db.fetchall(
            """SELECT * FROM bucket_transactions
               WHERE type = ?
               ORDER BY created_at DESC
               LIMIT ?""",
            (transaction_type.value, limit),
        )
        return [self._row_to_transaction(row) for row in rows]

    async def delete_transactions(self, bucket_id: str) -> int:
        """Delete all transactions for a bucket.

        Note: Should typically only be used for cleanup during tests
        or when retiring a satellite.

        Returns:
            Number of transactions deleted
        """
        result = await self._db.execute(
            "DELETE FROM bucket_transactions WHERE bucket_id = ?",
            (bucket_id,),
        )
        await self._db.commit()
        return result.rowcount

    # Allocation settings

    async def get_allocation_setting(self, key: str) -> Optional[float]:
        """Get an allocation setting value."""
        row = await self._db.fetchone(
            "SELECT value FROM allocation_settings WHERE key = ?",
            (key,),
        )
        return row["value"] if row else None

    async def set_allocation_setting(
        self, key: str, value: float, description: Optional[str] = None
    ) -> None:
        """Set an allocation setting value."""
        if description:
            await self._db.execute(
                """INSERT OR REPLACE INTO allocation_settings
                   (key, value, description)
                   VALUES (?, ?, ?)""",
                (key, value, description),
            )
        else:
            await self._db.execute(
                """UPDATE allocation_settings
                   SET value = ?
                   WHERE key = ?""",
                (value, key),
            )
        await self._db.commit()

    async def get_all_allocation_settings(self) -> Dict[str, float]:
        """Get all allocation settings as a dictionary."""
        rows = await self._db.fetchall("SELECT key, value FROM allocation_settings")
        return {row["key"]: row["value"] for row in rows}

    # Helper methods

    def _row_to_balance(self, row) -> BucketBalance:
        """Convert database row to BucketBalance model."""
        return BucketBalance(
            bucket_id=row["bucket_id"],
            currency=row["currency"],
            balance=row["balance"],
            last_updated=row["last_updated"],
        )

    def _row_to_transaction(self, row) -> BucketTransaction:
        """Convert database row to BucketTransaction model."""
        return BucketTransaction(
            id=row["id"],
            bucket_id=row["bucket_id"],
            type=TransactionType(row["type"]),
            amount=row["amount"],
            currency=row["currency"],
            description=row["description"],
            created_at=row["created_at"],
        )
