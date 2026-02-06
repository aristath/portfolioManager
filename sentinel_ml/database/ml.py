"""ML Database — Separate database for all ML tables.

All ML data (training samples, models, predictions, performance, regime state/models)
lives in data/ml.db, isolated from the main sentinel.db.

Per-model tables: one set of (models, predictions, performance) tables
for each model type: xgboost, ridge, rf, svr.
"""

import logging
import time
from pathlib import Path
from typing import Optional

import aiosqlite
import numpy as np
import pandas as pd

from sentinel_ml.ml_features import FEATURE_NAMES

logger = logging.getLogger(__name__)

MODEL_TYPES = ["xgboost", "ridge", "rf", "svr"]


class MLDatabase:
    """Dedicated database for ML tables. Singleton per path."""

    _instances: dict[str, "MLDatabase"] = {}
    _path: Path
    _connection: Optional[aiosqlite.Connection]

    def __new__(cls, path: str | None = None):
        if path is None:
            from sentinel_ml.paths import DATA_DIR

            path = str(DATA_DIR / "ml.db")

        if path not in cls._instances:
            instance = super().__new__(cls)
            instance._path = Path(path)
            instance._connection = None
            cls._instances[path] = instance

        return cls._instances[path]

    def __init__(self, path: str | None = None):
        pass

    @property
    def conn(self) -> aiosqlite.Connection:
        if not self._connection:
            raise RuntimeError("MLDatabase not connected. Call connect() first.")
        return self._connection

    async def connect(self) -> "MLDatabase":
        if self._connection is None:
            self._path.parent.mkdir(parents=True, exist_ok=True)
            self._connection = await aiosqlite.connect(self._path)
            self._connection.row_factory = aiosqlite.Row
            await self._connection.execute("PRAGMA journal_mode=WAL")
            await self._connection.execute("PRAGMA busy_timeout=30000")
            await self._init_schema()
        return self

    async def close(self):
        if self._connection:
            await self._connection.close()
            self._connection = None

    def remove_from_cache(self):
        path_str = str(self._path)
        if path_str in self._instances:
            del self._instances[path_str]

    # -------------------------------------------------------------------------
    # Schema
    # -------------------------------------------------------------------------

    async def _init_schema(self):
        # Training samples (shared, ghost columns removed)
        await self.conn.execute("""
        CREATE TABLE IF NOT EXISTS ml_training_samples (
            sample_id TEXT PRIMARY KEY,
            symbol TEXT NOT NULL,
            sample_date INTEGER NOT NULL,
            return_1d REAL, return_5d REAL, return_20d REAL, return_60d REAL,
            price_normalized REAL,
            volatility_10d REAL, volatility_30d REAL, atr_14d REAL,
            volume_normalized REAL, volume_trend REAL,
            rsi_14 REAL, macd REAL, bollinger_position REAL,
            sentiment_score REAL,
            country_agg_momentum REAL, country_agg_rsi REAL, country_agg_volatility REAL,
            industry_agg_momentum REAL, industry_agg_rsi REAL, industry_agg_volatility REAL,
            future_return REAL,
            prediction_horizon_days INTEGER,
            created_at INTEGER NOT NULL,
            UNIQUE(symbol, sample_date)
        )
        """)
        await self.conn.execute(
            "CREATE INDEX IF NOT EXISTS idx_ml_samples_symbol_date ON ml_training_samples(symbol, sample_date DESC)"
        )

        # Per-model tables
        for mt in MODEL_TYPES:
            await self.conn.execute(f"""
            CREATE TABLE IF NOT EXISTS ml_models_{mt} (
                symbol TEXT PRIMARY KEY,
                training_samples INTEGER,
                validation_rmse REAL,
                validation_mae REAL,
                validation_r2 REAL,
                last_trained_at INTEGER NOT NULL
            )
            """)

            await self.conn.execute(f"""
            CREATE TABLE IF NOT EXISTS ml_predictions_{mt} (
                prediction_id TEXT PRIMARY KEY,
                symbol TEXT NOT NULL,
                predicted_at INTEGER NOT NULL,
                features TEXT,
                predicted_return REAL,
                ml_score REAL,
                regime_score REAL,
                regime_dampening REAL,
                inference_time_ms REAL
            )
            """)
            await self.conn.execute(
                f"CREATE INDEX IF NOT EXISTS idx_ml_predictions_{mt}_symbol_date "
                f"ON ml_predictions_{mt}(symbol, predicted_at DESC)"
            )

            await self.conn.execute(f"""
            CREATE TABLE IF NOT EXISTS ml_performance_{mt} (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                symbol TEXT NOT NULL,
                tracked_at INTEGER NOT NULL,
                mean_absolute_error REAL,
                root_mean_squared_error REAL,
                prediction_bias REAL,
                drift_detected INTEGER DEFAULT 0,
                predictions_evaluated INTEGER DEFAULT 0,
                UNIQUE(symbol, tracked_at)
            )
            """)

        # Regime detection tables (ML-specific)
        await self.conn.execute("""
        CREATE TABLE IF NOT EXISTS regime_states (
            symbol TEXT NOT NULL,
            date TEXT NOT NULL,
            regime INTEGER NOT NULL,
            regime_name TEXT,
            confidence REAL,
            PRIMARY KEY (symbol, date)
        )
        """)
        await self.conn.execute("""
        CREATE TABLE IF NOT EXISTS regime_models (
            model_id TEXT PRIMARY KEY,
            symbols TEXT,
            n_states INTEGER,
            trained_at TEXT,
            model_params TEXT
        )
        """)
        await self.conn.execute("CREATE INDEX IF NOT EXISTS idx_regime_symbol_date ON regime_states(symbol, date DESC)")

        # ML scheduler tables
        await self.conn.execute(
            """
            CREATE TABLE IF NOT EXISTS ml_job_schedules (
                job_type TEXT PRIMARY KEY,
                interval_minutes INTEGER NOT NULL,
                market_timing INTEGER NOT NULL DEFAULT 0,
                interval_market_open_minutes INTEGER,
                description TEXT,
                category TEXT,
                last_run INTEGER NOT NULL DEFAULT 0
            )
            """
        )
        await self.conn.execute(
            """
            CREATE TABLE IF NOT EXISTS ml_job_history (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                job_id TEXT NOT NULL,
                job_type TEXT NOT NULL,
                status TEXT NOT NULL,
                error TEXT,
                duration_ms INTEGER,
                rows_affected INTEGER,
                executed_at INTEGER NOT NULL
            )
            """
        )
        await self.conn.execute(
            "CREATE INDEX IF NOT EXISTS idx_ml_job_history_type_executed ON ml_job_history(job_type, executed_at DESC)"
        )
        await self.conn.execute(
            """
            CREATE TABLE IF NOT EXISTS ml_service_settings (
                key TEXT PRIMARY KEY,
                value TEXT
            )
            """
        )

        await self.conn.commit()
        await self._migrate_ml_job_schedule_columns()

    async def _migrate_ml_job_schedule_columns(self) -> None:
        cursor = await self.conn.execute("PRAGMA table_info(ml_job_schedules)")
        cols = {row["name"] for row in await cursor.fetchall()}
        if "description" not in cols:
            await self.conn.execute("ALTER TABLE ml_job_schedules ADD COLUMN description TEXT")
        if "category" not in cols:
            await self.conn.execute("ALTER TABLE ml_job_schedules ADD COLUMN category TEXT")
        await self.conn.commit()

    # -------------------------------------------------------------------------
    # ML Scheduler state
    # -------------------------------------------------------------------------

    async def seed_default_job_schedules(self) -> None:
        defaults = [
            ("analytics:regime", 1440, "Train regime detection model", "analytics"),
            ("ml:retrain", 10080, "Retrain ML models for all ML-enabled securities", "ml"),
            ("ml:monitor", 1440, "Monitor ML performance for all ML-enabled securities", "ml"),
        ]
        for job_type, interval, description, category in defaults:
            await self.conn.execute(
                """
                INSERT OR IGNORE INTO ml_job_schedules
                    (job_type, interval_minutes, market_timing, description, category, last_run)
                VALUES (?, ?, 0, ?, ?, 0)
                """,
                (job_type, interval, description, category),
            )
            await self.conn.execute(
                "UPDATE ml_job_schedules SET description = ?, category = ? WHERE job_type = ?",
                (description, category, job_type),
            )
        await self.conn.commit()

    async def get_job_schedules(self) -> list[dict]:
        cursor = await self.conn.execute("SELECT * FROM ml_job_schedules ORDER BY job_type")
        return [dict(row) for row in await cursor.fetchall()]

    async def get_job_schedule(self, job_type: str) -> dict | None:
        cursor = await self.conn.execute("SELECT * FROM ml_job_schedules WHERE job_type = ?", (job_type,))
        row = await cursor.fetchone()
        return dict(row) if row else None

    async def upsert_job_schedule(
        self,
        job_type: str,
        interval_minutes: int | None = None,
        interval_market_open_minutes: int | None = None,
        market_timing: int | None = None,
    ) -> None:
        existing = await self.get_job_schedule(job_type)
        if not existing:
            await self.conn.execute(
                """
                INSERT INTO ml_job_schedules (
                    job_type, interval_minutes, market_timing, interval_market_open_minutes,
                    description, category, last_run
                )
                VALUES (?, ?, ?, ?, '', '', 0)
                """,
                (job_type, interval_minutes or 60, market_timing or 0, interval_market_open_minutes),
            )
        else:
            await self.conn.execute(
                """
                UPDATE ml_job_schedules
                SET interval_minutes = ?, interval_market_open_minutes = ?, market_timing = ?
                WHERE job_type = ?
                """,
                (
                    interval_minutes if interval_minutes is not None else existing["interval_minutes"],
                    interval_market_open_minutes
                    if interval_market_open_minutes is not None
                    else existing.get("interval_market_open_minutes"),
                    market_timing if market_timing is not None else existing["market_timing"],
                    job_type,
                ),
            )
        await self.conn.commit()

    async def mark_job_completed(self, job_type: str) -> None:
        await self.conn.execute(
            "UPDATE ml_job_schedules SET last_run = ? WHERE job_type = ?", (int(time.time()), job_type)
        )
        await self.conn.commit()

    async def mark_job_failed(self, job_type: str) -> None:
        await self.conn.execute(
            "UPDATE ml_job_schedules SET last_run = ? WHERE job_type = ?", (int(time.time()), job_type)
        )
        await self.conn.commit()

    async def log_job_execution(
        self,
        job_id: str,
        job_type: str,
        status: str,
        error: str | None,
        duration_ms: int,
        rows_affected: int = 0,
    ) -> None:
        await self.conn.execute(
            """
            INSERT INTO ml_job_history (job_id, job_type, status, error, duration_ms, rows_affected, executed_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            """,
            (job_id, job_type, status, error, duration_ms, rows_affected, int(time.time())),
        )
        await self.conn.commit()

    async def get_job_history(self, limit: int = 50) -> list[dict]:
        cursor = await self.conn.execute("SELECT * FROM ml_job_history ORDER BY executed_at DESC LIMIT ?", (limit,))
        return [dict(row) for row in await cursor.fetchall()]

    # -------------------------------------------------------------------------
    # Service settings
    # -------------------------------------------------------------------------

    async def get_service_setting(self, key: str, default: str | None = None) -> str | None:
        cursor = await self.conn.execute("SELECT value FROM ml_service_settings WHERE key = ?", (key,))
        row = await cursor.fetchone()
        if row is None:
            return default
        return row["value"]

    async def set_service_setting(self, key: str, value: str) -> None:
        await self.conn.execute(
            "INSERT OR REPLACE INTO ml_service_settings (key, value) VALUES (?, ?)",
            (key, value),
        )
        await self.conn.commit()

    async def get_job_history_for_type(self, job_type: str, limit: int = 50) -> list[dict]:
        cursor = await self.conn.execute(
            "SELECT * FROM ml_job_history WHERE job_type = ? ORDER BY executed_at DESC LIMIT ?",
            (job_type, limit),
        )
        return [dict(row) for row in await cursor.fetchall()]

    # -------------------------------------------------------------------------
    # Training Samples
    # -------------------------------------------------------------------------

    async def store_training_samples(self, df: pd.DataFrame) -> None:
        """Batch INSERT OR REPLACE training samples from a DataFrame."""
        if len(df) == 0:
            return

        db_columns = (
            ["sample_id", "symbol", "sample_date"]
            + list(FEATURE_NAMES)
            + ["future_return", "prediction_horizon_days", "created_at"]
        )

        sql = f"""
            INSERT OR REPLACE INTO ml_training_samples
            ({", ".join(db_columns)})
            VALUES ({", ".join(["?" for _ in db_columns])})
        """  # noqa: S608

        for _, row in df.iterrows():
            values = []
            for col in db_columns:
                val = row.get(col, 0.0)
                # Convert numpy types to Python native
                if hasattr(val, "item") and val is not None:
                    val = val.item()
                if val is None:
                    val = 0.0
                elif np.isscalar(val) and bool(pd.isna(val)):
                    val = 0.0
                values.append(val)
            await self.conn.execute(sql, tuple(values))

        await self.conn.commit()

    async def load_training_data(self, symbol: str) -> tuple[np.ndarray, np.ndarray]:
        """Load training data for a symbol as (X, y) arrays."""
        cursor = await self.conn.execute(
            """SELECT * FROM ml_training_samples
               WHERE symbol = ? AND future_return IS NOT NULL
               ORDER BY sample_date DESC""",
            (symbol,),
        )
        rows = await cursor.fetchall()

        if not rows:
            return np.array([]), np.array([])

        from sentinel_ml.ml_features import DEFAULT_FEATURES

        df = pd.DataFrame([dict(row) for row in rows])

        for col in FEATURE_NAMES:
            if col not in df.columns:
                df[col] = DEFAULT_FEATURES.get(col, 0.0)
            else:
                df[col] = df[col].fillna(DEFAULT_FEATURES.get(col, 0.0))

        X = df[list(FEATURE_NAMES)].values.astype(np.float32)
        y = df["future_return"].values.astype(np.float32)

        valid_mask = np.all(np.isfinite(X), axis=1) & np.isfinite(y)
        return X[valid_mask], y[valid_mask]

    async def get_symbols_with_sufficient_data(self, min_samples: int) -> dict[str, int]:
        """Get symbols with at least min_samples training rows."""
        cursor = await self.conn.execute(
            """SELECT symbol, COUNT(*) as sample_count
               FROM ml_training_samples
               WHERE future_return IS NOT NULL
               GROUP BY symbol
               HAVING sample_count >= ?
               ORDER BY sample_count DESC""",
            (min_samples,),
        )
        rows = await cursor.fetchall()
        return {row["symbol"]: row["sample_count"] for row in rows}

    async def get_sample_count(self, symbol: str) -> int:
        """Get training sample count for a symbol."""
        cursor = await self.conn.execute(
            "SELECT COUNT(*) as count FROM ml_training_samples WHERE symbol = ? AND future_return IS NOT NULL",
            (symbol,),
        )
        row = await cursor.fetchone()
        return row["count"] if row else 0

    # -------------------------------------------------------------------------
    # Predictions (per-model)
    # -------------------------------------------------------------------------

    async def store_prediction(
        self,
        model_type: str,
        prediction_id: str,
        symbol: str,
        predicted_at: int,
        features: str | None,
        predicted_return: float,
        ml_score: float,
        regime_score: float | None,
        regime_dampening: float | None,
        inference_time_ms: float,
    ) -> None:
        """Insert a prediction into the per-model table."""
        sql = (
            f"INSERT OR REPLACE INTO ml_predictions_{model_type} "  # noqa: S608
            "(prediction_id, symbol, predicted_at, features, "
            "predicted_return, ml_score, regime_score, regime_dampening, "
            "inference_time_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
        )
        await self.conn.execute(
            sql,
            (
                prediction_id,
                symbol,
                predicted_at,
                features,
                predicted_return,
                ml_score,
                regime_score,
                regime_dampening,
                inference_time_ms,
            ),
        )
        await self.conn.commit()

    async def get_prediction_as_of(self, model_type: str, symbol: str, as_of_ts: int) -> dict | None:
        """Get most recent prediction for symbol as of timestamp."""
        sql = (
            f"SELECT * FROM ml_predictions_{model_type} "  # noqa: S608
            "WHERE symbol = ? AND predicted_at <= ? "
            "ORDER BY predicted_at DESC LIMIT 1"
        )
        cursor = await self.conn.execute(
            sql,
            (symbol, as_of_ts),
        )
        row = await cursor.fetchone()
        return dict(row) if row else None

    async def get_latest_prediction(self, model_type: str, symbol: str) -> dict | None:
        """Get most recent prediction for symbol."""
        sql = (
            f"SELECT * FROM ml_predictions_{model_type} "  # noqa: S608
            "WHERE symbol = ? "
            "ORDER BY predicted_at DESC LIMIT 1"
        )
        cursor = await self.conn.execute(sql, (symbol,))
        row = await cursor.fetchone()
        return dict(row) if row else None

    async def get_all_predictions_history(self, model_type: str) -> list[dict]:
        """Get all predictions for a model type ordered by symbol and date."""
        sql = (
            f"SELECT symbol, predicted_return, predicted_at FROM ml_predictions_{model_type} "  # noqa: S608
            "ORDER BY symbol, predicted_at"
        )
        cursor = await self.conn.execute(sql)
        return [dict(row) for row in await cursor.fetchall()]

    # -------------------------------------------------------------------------
    # Model Records (per-model)
    # -------------------------------------------------------------------------

    async def update_model_record(
        self,
        model_type: str,
        symbol: str,
        training_samples: int,
        metrics: dict,
    ) -> None:
        """Insert or update model record."""
        sql = (
            f"INSERT OR REPLACE INTO ml_models_{model_type} "  # noqa: S608
            "(symbol, training_samples, validation_rmse, validation_mae, "
            "validation_r2, last_trained_at) VALUES (?, ?, ?, ?, ?, ?)"
        )
        await self.conn.execute(
            sql,
            (
                symbol,
                training_samples,
                metrics["validation_rmse"],
                metrics["validation_mae"],
                metrics["validation_r2"],
                int(time.time()),
            ),
        )
        await self.conn.commit()

    async def get_model_status(self, model_type: str) -> list[dict]:
        """Get all model records for a model type."""
        sql = f"SELECT * FROM ml_models_{model_type} ORDER BY last_trained_at DESC"  # noqa: S608
        cursor = await self.conn.execute(sql)
        return [dict(row) for row in await cursor.fetchall()]

    async def get_all_model_status(self) -> dict[str, list[dict]]:
        """Get model records for all model types."""
        result = {}
        for mt in MODEL_TYPES:
            result[mt] = await self.get_model_status(mt)
        return result

    # -------------------------------------------------------------------------
    # Performance (per-model)
    # -------------------------------------------------------------------------

    async def store_performance_metrics(
        self,
        model_type: str,
        symbol: str,
        tracked_at: int,
        metrics: dict,
    ) -> None:
        """Store performance metrics for a model type."""
        sql = (
            f"INSERT OR REPLACE INTO ml_performance_{model_type} "  # noqa: S608
            "(symbol, tracked_at, mean_absolute_error, root_mean_squared_error, "
            "prediction_bias, drift_detected, predictions_evaluated) "
            "VALUES (?, ?, ?, ?, ?, 0, ?)"
        )
        await self.conn.execute(
            sql,
            (
                symbol,
                tracked_at,
                metrics.get("mean_absolute_error"),
                metrics.get("root_mean_squared_error"),
                metrics.get("prediction_bias"),
                metrics.get("predictions_evaluated", 0),
            ),
        )
        await self.conn.commit()

    # -------------------------------------------------------------------------
    # Deletion
    # -------------------------------------------------------------------------

    async def delete_all_data(self) -> None:
        """Truncate all ML tables."""
        await self.conn.execute("DELETE FROM ml_training_samples")
        for mt in MODEL_TYPES:
            await self.conn.execute(f"DELETE FROM ml_models_{mt}")  # noqa: S608
            await self.conn.execute(f"DELETE FROM ml_predictions_{mt}")  # noqa: S608
            await self.conn.execute(f"DELETE FROM ml_performance_{mt}")  # noqa: S608
        await self.conn.execute("DELETE FROM regime_states")
        await self.conn.execute("DELETE FROM regime_models")
        await self.conn.commit()

    async def delete_symbol_data(self, symbol: str) -> None:
        """Delete all data for one symbol from all tables."""
        await self.conn.execute("DELETE FROM ml_training_samples WHERE symbol = ?", (symbol,))
        for mt in MODEL_TYPES:
            await self.conn.execute(f"DELETE FROM ml_models_{mt} WHERE symbol = ?", (symbol,))  # noqa: S608
            await self.conn.execute(f"DELETE FROM ml_predictions_{mt} WHERE symbol = ?", (symbol,))  # noqa: S608
            await self.conn.execute(f"DELETE FROM ml_performance_{mt} WHERE symbol = ?", (symbol,))  # noqa: S608
        await self.conn.execute("DELETE FROM regime_states WHERE symbol = ?", (symbol,))
        await self.conn.commit()
