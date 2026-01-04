#!/usr/bin/env python3
"""
Migrate data from Python app databases to Go app databases.

This script migrates data from the legacy Python application's database structure
to the new Go application's 8-database architecture.

Usage:
    python scripts/migrate_python_to_go.py [--python-data-dir PATH] [--go-data-dir PATH] [--dry-run]

The script will:
1. Create backups of all databases
2. Initialize Go database schemas
3. Migrate data with schema transformations
4. Verify migration success
"""

import argparse
import asyncio
import logging
import shutil
import sqlite3
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'migration_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)


class DatabaseMigrator:
    """Handles migration from Python databases to Go databases."""

    def __init__(self, python_data_dir: Path, go_data_dir: Path, dry_run: bool = False):
        self.python_data_dir = Path(python_data_dir)
        self.go_data_dir = Path(go_data_dir)
        self.dry_run = dry_run

        # Ensure directories exist
        self.go_data_dir.mkdir(parents=True, exist_ok=True)
        self.backup_dir = self.go_data_dir / "backups" / f"migration_{datetime.now().strftime('%Y%m%d_%H%M%S')}"

        if not dry_run:
            self.backup_dir.mkdir(parents=True, exist_ok=True)

    def backup_database(self, db_path: Path) -> bool:
        """Backup a database file."""
        if not db_path.exists():
            logger.warning(f"Database not found: {db_path}")
            return False

        if self.dry_run:
            logger.info(f"[DRY RUN] Would backup: {db_path} -> {self.backup_dir / db_path.name}")
            return True

        try:
            shutil.copy2(db_path, self.backup_dir / db_path.name)
            logger.info(f"Backed up: {db_path.name}")
            return True
        except Exception as e:
            logger.error(f"Failed to backup {db_path}: {e}")
            return False

    def get_table_schema(self, conn: sqlite3.Connection, table_name: str) -> List[Tuple[str, str]]:
        """Get column names and types for a table."""
        cursor = conn.execute(f"PRAGMA table_info({table_name})")
        return [(row[1], row[2]) for row in cursor.fetchall()]

    def table_exists(self, conn: sqlite3.Connection, table_name: str) -> bool:
        """Check if a table exists."""
        cursor = conn.execute(
            "SELECT name FROM sqlite_master WHERE type='table' AND name=?",
            (table_name,)
        )
        return cursor.fetchone() is not None

    def get_row_count(self, conn: sqlite3.Connection, table_name: str) -> int:
        """Get row count for a table."""
        if not self.table_exists(conn, table_name):
            return 0
        cursor = conn.execute(f"SELECT COUNT(*) FROM {table_name}")
        return cursor.fetchone()[0]

    def migrate_universe(self):
        """Migrate securities and groups from config.db to universe.db"""
        logger.info("=" * 60)
        logger.info("Migrating Universe Database")
        logger.info("=" * 60)

        python_config = self.python_data_dir / "config.db"
        go_universe = self.go_data_dir / "universe.db"

        if not python_config.exists():
            logger.warning(f"Python config.db not found: {python_config}")
            return

        # Initialize Go universe schema
        if not self.dry_run:
            self._init_universe_schema(go_universe)

        python_conn = sqlite3.connect(python_config)
        go_conn = sqlite3.connect(go_universe)

        try:
            # Migrate securities
            if self.table_exists(python_conn, "securities"):
                python_cols = [col[0] for col in self.get_table_schema(python_conn, "securities")]
                go_cols = [col[0] for col in self.get_table_schema(go_conn, "securities")]

                # Map Python columns to Go columns
                col_mapping = {
                    "symbol": "symbol",
                    "yahoo_symbol": "yahoo_symbol",
                    "isin": "isin",
                    "name": "name",
                    "product_type": "product_type",
                    "industry": "industry",
                    "country": "country",
                    "geography": "country",  # Python uses 'geography', Go uses 'country'
                    "fullExchangeName": "fullExchangeName",
                    "priority_multiplier": "priority_multiplier",
                    "min_lot": "min_lot",
                    "active": "active",
                    "allow_buy": "allow_buy",
                    "allow_sell": "allow_sell",
                    "currency": "currency",
                    "last_synced": "last_synced",
                    "min_portfolio_target": "min_portfolio_target",
                    "max_portfolio_target": "max_portfolio_target",
                    "bucket_id": "bucket_id",
                }

                # Select columns that exist in both
                select_cols = [col for col in python_cols if col in col_mapping]
                insert_cols = [col_mapping[col] for col in select_cols if col_mapping[col] in go_cols]

                # Add default values for required Go columns
                now = datetime.now().isoformat()
                if "created_at" in go_cols and "created_at" not in insert_cols:
                    insert_cols.append("created_at")
                if "updated_at" in go_cols and "updated_at" not in insert_cols:
                    insert_cols.append("updated_at")

                select_sql = f"SELECT {', '.join(select_cols)} FROM securities"
                python_rows = python_conn.execute(select_sql).fetchall()

                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    placeholders = ', '.join(['?' for _ in insert_cols])
                    insert_sql = f"INSERT OR REPLACE INTO securities ({', '.join(insert_cols)}) VALUES ({placeholders})"

                    for row in python_rows:
                        values = list(row)
                        # Convert boolean TEXT values to INTEGER for Go schema
                        row_dict = dict(zip(select_cols, row))
                        for i, col in enumerate(select_cols):
                            if col in ['active', 'allow_buy', 'allow_sell']:
                                # Convert Python boolean/TEXT to INTEGER (0 or 1)
                                val = row_dict[col]
                                if isinstance(val, bool):
                                    values[i] = 1 if val else 0
                                elif isinstance(val, str):
                                    values[i] = 1 if val.lower() in ['true', '1', 'yes', 't'] else 0
                                elif val is None:
                                    values[i] = 1  # Default to active
                                else:
                                    values[i] = 1 if val else 0

                        # Add default timestamps if needed
                        if "created_at" in insert_cols and "created_at" not in select_cols:
                            values.append(now)
                        if "updated_at" in insert_cols and "updated_at" not in select_cols:
                            values.append(now)
                        go_cursor.execute(insert_sql, values)

                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} securities")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} securities")

            # Migrate country_groups
            if self.table_exists(python_conn, "country_groups"):
                python_rows = python_conn.execute("SELECT * FROM country_groups").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    for row in python_rows:
                        go_cursor.execute(
                            "INSERT OR REPLACE INTO country_groups (group_name, country_name, created_at, updated_at) VALUES (?, ?, ?, ?)",
                            (row[0], row[1], datetime.now().isoformat(), datetime.now().isoformat())
                        )
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} country groups")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} country groups")

            # Migrate industry_groups
            if self.table_exists(python_conn, "industry_groups"):
                python_rows = python_conn.execute("SELECT * FROM industry_groups").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    for row in python_rows:
                        go_cursor.execute(
                            "INSERT OR REPLACE INTO industry_groups (group_name, industry_name, created_at, updated_at) VALUES (?, ?, ?, ?)",
                            (row[0], row[1], datetime.now().isoformat(), datetime.now().isoformat())
                        )
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} industry groups")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} industry groups")

        finally:
            python_conn.close()
            go_conn.close()

    def migrate_portfolio(self):
        """Migrate positions, scores, and snapshots to portfolio.db"""
        logger.info("=" * 60)
        logger.info("Migrating Portfolio Database")
        logger.info("=" * 60)

        python_state = self.python_data_dir / "state.db"
        python_calculations = self.python_data_dir / "calculations.db"
        python_snapshots = self.python_data_dir / "snapshots.db"
        go_portfolio = self.go_data_dir / "portfolio.db"

        if not python_state.exists():
            logger.warning(f"Python state.db not found: {python_state}")
            return

        # Initialize Go portfolio schema
        if not self.dry_run:
            self._init_portfolio_schema(go_portfolio)

        python_state_conn = sqlite3.connect(python_state)
        python_calc_conn = sqlite3.connect(python_calculations) if python_calculations.exists() else None
        python_snap_conn = sqlite3.connect(python_snapshots) if python_snapshots.exists() else None
        go_conn = sqlite3.connect(go_portfolio)

        try:
            # Migrate positions
            if self.table_exists(python_state_conn, "positions"):
                python_rows = python_state_conn.execute("SELECT * FROM positions").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_state_conn, "positions")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO positions (
                                symbol, quantity, avg_price, current_price, currency,
                                currency_rate, market_value_eur, cost_basis_eur,
                                unrealized_pnl, unrealized_pnl_pct, last_updated,
                                first_bought, last_sold, isin, bucket_id
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("symbol"),
                            row_dict.get("quantity", 0),
                            row_dict.get("avg_price", 0),
                            row_dict.get("current_price"),
                            row_dict.get("currency"),
                            row_dict.get("currency_rate", 1.0),
                            row_dict.get("market_value_eur"),
                            row_dict.get("cost_basis_eur"),
                            row_dict.get("unrealized_pnl"),
                            row_dict.get("unrealized_pnl_pct"),
                            row_dict.get("last_updated"),
                            row_dict.get("first_bought"),
                            row_dict.get("last_sold"),
                            row_dict.get("isin"),
                            row_dict.get("bucket_id", "core"),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} positions")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} positions")

            # Migrate scores
            scores_migrated = False
            if self.table_exists(python_state_conn, "scores"):
                python_rows = python_state_conn.execute("SELECT * FROM scores").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_state_conn, "scores")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO scores (
                                symbol, total_score, quality_score, opportunity_score,
                                analyst_score, allocation_fit_score, volatility, cagr_score,
                                consistency_score, history_years, technical_score,
                                fundamental_score, last_updated
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("symbol"),
                            row_dict.get("total_score", 0),
                            row_dict.get("quality_score"),
                            row_dict.get("opportunity_score"),
                            row_dict.get("analyst_score"),
                            row_dict.get("allocation_fit_score"),
                            row_dict.get("volatility"),
                            row_dict.get("cagr_score"),
                            row_dict.get("consistency_score"),
                            row_dict.get("history_years"),
                            row_dict.get("technical_score"),
                            row_dict.get("fundamental_score"),
                            row_dict.get("last_updated", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} scores from state.db")
                    scores_migrated = True
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} scores from state.db")

            # Try calculations.db for scores if not found in state.db
            if python_calc_conn and not scores_migrated and self.table_exists(python_calc_conn, "scores"):
                python_rows = python_calc_conn.execute("SELECT * FROM scores").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_calc_conn, "scores")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO scores (
                                symbol, total_score, quality_score, opportunity_score,
                                analyst_score, allocation_fit_score, volatility, cagr_score,
                                consistency_score, history_years, technical_score,
                                fundamental_score, last_updated
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("symbol"),
                            row_dict.get("total_score", 0),
                            row_dict.get("quality_score"),
                            row_dict.get("opportunity_score"),
                            row_dict.get("analyst_score"),
                            row_dict.get("allocation_fit_score"),
                            row_dict.get("volatility"),
                            row_dict.get("cagr_score"),
                            row_dict.get("consistency_score"),
                            row_dict.get("history_years"),
                            row_dict.get("technical_score"),
                            row_dict.get("fundamental_score"),
                            row_dict.get("last_updated", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} scores from calculations.db")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} scores from calculations.db")

            # Migrate portfolio snapshots
            if python_snap_conn and self.table_exists(python_snap_conn, "portfolio_snapshots"):
                python_rows = python_snap_conn.execute("SELECT * FROM portfolio_snapshots").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_snap_conn, "portfolio_snapshots")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO portfolio_snapshots (
                                snapshot_date, total_value, cash_balance, invested_value,
                                total_pnl, total_pnl_pct, position_count, bucket_id,
                                snapshot_json, created_at
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("snapshot_date"),
                            row_dict.get("total_value", 0),
                            row_dict.get("cash_balance", 0),
                            row_dict.get("invested_value", 0),
                            row_dict.get("total_pnl"),
                            row_dict.get("total_pnl_pct"),
                            row_dict.get("position_count", 0),
                            row_dict.get("bucket_id", "core"),
                            row_dict.get("snapshot_json"),
                            row_dict.get("created_at", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} portfolio snapshots")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} portfolio snapshots")

        finally:
            python_state_conn.close()
            if python_calc_conn:
                python_calc_conn.close()
            if python_snap_conn:
                python_snap_conn.close()
            go_conn.close()

    def migrate_ledger(self):
        """Migrate trades, cash flows, and dividends to ledger.db"""
        logger.info("=" * 60)
        logger.info("Migrating Ledger Database")
        logger.info("=" * 60)

        python_ledger = self.python_data_dir / "ledger.db"
        python_dividends = self.python_data_dir / "dividends.db"
        go_ledger = self.go_data_dir / "ledger.db"

        if not python_ledger.exists():
            logger.warning(f"Python ledger.db not found: {python_ledger}")
            return

        # Initialize Go ledger schema
        if not self.dry_run:
            self._init_ledger_schema(go_ledger)

        python_ledger_conn = sqlite3.connect(python_ledger)
        python_div_conn = sqlite3.connect(python_dividends) if python_dividends.exists() else None
        go_conn = sqlite3.connect(go_ledger)

        try:
            # Migrate trades
            if self.table_exists(python_ledger_conn, "trades"):
                python_rows = python_ledger_conn.execute("SELECT * FROM trades").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_ledger_conn, "trades")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        # Normalize side value to 'buy' or 'sell'
                        side = row_dict.get("side", "buy")
                        if isinstance(side, str):
                            side_lower = side.lower()
                            if side_lower in ['buy', 'b', 'purchase']:
                                side = 'buy'
                            elif side_lower in ['sell', 's', 'sale']:
                                side = 'sell'
                            else:
                                side = 'buy'  # Default
                        else:
                            side = 'buy'

                        go_cursor.execute("""
                            INSERT OR REPLACE INTO trades (
                                id, symbol, isin, side, quantity, price, executed_at,
                                order_id, currency, value_eur, source, bucket_id, mode, created_at
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("id"),
                            row_dict.get("symbol") or "",
                            row_dict.get("isin"),
                            side,
                            row_dict.get("quantity", 0),
                            row_dict.get("price", 0),
                            row_dict.get("executed_at") or datetime.now().isoformat(),
                            row_dict.get("order_id"),
                            row_dict.get("currency") or "EUR",  # Default to EUR if missing
                            row_dict.get("value_eur") or 0,
                            row_dict.get("source", "manual"),
                            row_dict.get("bucket_id", "core"),
                            row_dict.get("mode", "normal"),
                            row_dict.get("created_at", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} trades")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} trades")

            # Migrate cash_flows
            if self.table_exists(python_ledger_conn, "cash_flows"):
                python_rows = python_ledger_conn.execute("SELECT * FROM cash_flows").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_ledger_conn, "cash_flows")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        # Map Python transaction_type to Go flow_type
                        flow_type = row_dict.get("transaction_type", "deposit")
                        if flow_type not in ["deposit", "withdrawal", "fee", "dividend", "interest"]:
                            flow_type = "deposit"  # Default

                        go_cursor.execute("""
                            INSERT OR REPLACE INTO cash_flows (
                                id, flow_type, amount, currency, amount_eur,
                                description, executed_at, bucket_id, created_at
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("id"),
                            flow_type,
                            row_dict.get("amount", 0),
                            row_dict.get("currency"),
                            row_dict.get("amount_eur", 0),
                            row_dict.get("description"),
                            row_dict.get("date") or row_dict.get("executed_at"),
                            row_dict.get("bucket_id", "core"),
                            row_dict.get("created_at", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} cash flows")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} cash flows")

            # Migrate dividend_history
            if python_div_conn and self.table_exists(python_div_conn, "dividend_history"):
                python_rows = python_div_conn.execute("SELECT * FROM dividend_history").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_div_conn, "dividend_history")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO dividend_history (
                                id, symbol, isin, payment_date, ex_date,
                                amount_per_share, shares_held, total_amount, currency,
                                total_amount_eur, drip_enabled, reinvested_shares,
                                bucket_id, created_at
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("id"),
                            row_dict.get("symbol"),
                            row_dict.get("isin"),
                            row_dict.get("payment_date"),
                            row_dict.get("ex_date"),
                            row_dict.get("amount_per_share", 0),
                            row_dict.get("shares_held", 0),
                            row_dict.get("total_amount", 0),
                            row_dict.get("currency"),
                            row_dict.get("total_amount_eur", 0),
                            row_dict.get("drip_enabled", 0),
                            row_dict.get("reinvested_shares"),
                            row_dict.get("bucket_id", "core"),
                            row_dict.get("created_at", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} dividend history records")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} dividend history records")

        finally:
            python_ledger_conn.close()
            if python_div_conn:
                python_div_conn.close()
            go_conn.close()

    def migrate_satellites(self):
        """Migrate satellites/buckets data"""
        logger.info("=" * 60)
        logger.info("Migrating Satellites Database")
        logger.info("=" * 60)

        python_satellites = self.python_data_dir / "satellites.db"
        go_satellites = self.go_data_dir / "satellites.db"

        if not python_satellites.exists():
            logger.warning(f"Python satellites.db not found: {python_satellites}")
            return

        # Initialize Go satellites schema
        if not self.dry_run:
            self._init_satellites_schema(go_satellites)

        python_conn = sqlite3.connect(python_satellites)
        go_conn = sqlite3.connect(go_satellites)

        try:
            # Migrate buckets
            if self.table_exists(python_conn, "buckets"):
                python_rows = python_conn.execute("SELECT * FROM buckets").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, "buckets")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        now = datetime.now().isoformat()
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO buckets (
                                id, name, type, notes, target_pct, min_pct, max_pct,
                                consecutive_losses, max_consecutive_losses, high_water_mark,
                                high_water_mark_date, loss_streak_paused_at, status,
                                created_at, updated_at
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("id"),
                            row_dict.get("name"),
                            row_dict.get("type"),
                            row_dict.get("notes"),
                            row_dict.get("target_pct"),
                            row_dict.get("min_pct"),
                            row_dict.get("max_pct"),
                            row_dict.get("consecutive_losses", 0),
                            row_dict.get("max_consecutive_losses", 5),
                            row_dict.get("high_water_mark", 0),
                            row_dict.get("high_water_mark_date"),
                            row_dict.get("loss_streak_paused_at"),
                            row_dict.get("status", "active"),
                            row_dict.get("created_at", now),
                            row_dict.get("updated_at", now),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} buckets")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} buckets")

            # Migrate bucket_balances
            if self.table_exists(python_conn, "bucket_balances"):
                python_rows = python_conn.execute("SELECT * FROM bucket_balances").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, "bucket_balances")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO bucket_balances (
                                bucket_id, currency, balance, last_updated
                            ) VALUES (?, ?, ?, ?)
                        """, (
                            row_dict.get("bucket_id"),
                            row_dict.get("currency"),
                            row_dict.get("balance", 0),
                            row_dict.get("last_updated", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} bucket balances")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} bucket balances")

            # Migrate satellite_settings
            if self.table_exists(python_conn, "satellite_settings"):
                python_rows = python_conn.execute("SELECT * FROM satellite_settings").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, "satellite_settings")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO satellite_settings (
                                satellite_id, preset, risk_appetite, hold_duration,
                                entry_style, position_spread, profit_taking,
                                trailing_stops, follow_regime, auto_harvest,
                                pause_high_volatility, dividend_handling,
                                risk_free_rate, sortino_mar, evaluation_period_days,
                                volatility_window
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("satellite_id"),
                            row_dict.get("preset"),
                            row_dict.get("risk_appetite", 0.5),
                            row_dict.get("hold_duration", 0.5),
                            row_dict.get("entry_style", 0.5),
                            row_dict.get("position_spread", 0.5),
                            row_dict.get("profit_taking", 0.5),
                            row_dict.get("trailing_stops", 0),
                            row_dict.get("follow_regime", 0),
                            row_dict.get("auto_harvest", 0),
                            row_dict.get("pause_high_volatility", 0),
                            row_dict.get("dividend_handling", "reinvest_same"),
                            row_dict.get("risk_free_rate", 0.035),
                            row_dict.get("sortino_mar", 0.05),
                            row_dict.get("evaluation_period_days", 90),
                            row_dict.get("volatility_window", 60),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} satellite settings")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} satellite settings")

        finally:
            python_conn.close()
            go_conn.close()

    def migrate_agents(self):
        """Migrate planning/agent data from planner.db to agents.db"""
        logger.info("=" * 60)
        logger.info("Migrating Agents Database")
        logger.info("=" * 60)

        python_planner = self.python_data_dir / "planner.db"
        go_agents = self.go_data_dir / "agents.db"

        if not python_planner.exists():
            logger.warning(f"Python planner.db not found: {python_planner}")
            return

        # Initialize Go agents schema
        if not self.dry_run:
            self._init_agents_schema(go_agents)

        python_conn = sqlite3.connect(python_planner)
        go_conn = sqlite3.connect(go_agents)

        try:
            # Migrate sequences
            if self.table_exists(python_conn, "sequences"):
                python_rows = python_conn.execute("SELECT * FROM sequences").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, "sequences")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO sequences (
                                sequence_hash, portfolio_hash, priority, sequence_json,
                                depth, pattern_type, completed, evaluated_at, created_at
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("sequence_hash"),
                            row_dict.get("portfolio_hash"),
                            row_dict.get("priority", 0),
                            row_dict.get("sequence_json"),
                            row_dict.get("depth", 1),
                            row_dict.get("pattern_type"),
                            row_dict.get("completed", 0),
                            row_dict.get("evaluated_at"),
                            row_dict.get("created_at", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} sequences")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} sequences")

            # Migrate evaluations
            if self.table_exists(python_conn, "evaluations"):
                python_rows = python_conn.execute("SELECT * FROM evaluations").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, "evaluations")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO evaluations (
                                sequence_hash, portfolio_hash, end_score, breakdown_json,
                                end_cash, end_context_positions_json, div_score,
                                total_value, evaluated_at
                            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                        """, (
                            row_dict.get("sequence_hash"),
                            row_dict.get("portfolio_hash"),
                            row_dict.get("end_score", 0),
                            row_dict.get("breakdown_json"),
                            row_dict.get("end_cash", 0),
                            row_dict.get("end_context_positions_json"),
                            row_dict.get("div_score", 0),
                            row_dict.get("total_value", 0),
                            row_dict.get("evaluated_at", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} evaluations")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} evaluations")

        finally:
            python_conn.close()
            go_conn.close()

    def migrate_history(self):
        """Migrate historical price data from per-symbol databases to history.db"""
        logger.info("=" * 60)
        logger.info("Migrating History Database")
        logger.info("=" * 60)

        python_history_dir = self.python_data_dir / "history"
        go_history = self.go_data_dir / "history.db"

        if not python_history_dir.exists():
            logger.warning(f"Python history directory not found: {python_history_dir}")
            return

        # Initialize Go history schema
        if not self.dry_run:
            self._init_history_schema(go_history)

        go_conn = sqlite3.connect(go_history)

        try:
            history_files = list(python_history_dir.glob("*.db"))
            total_rows = 0

            for history_file in history_files:
                symbol = history_file.stem.replace("_", ".")  # Convert AAPL_US to AAPL.US
                python_conn = sqlite3.connect(history_file)

                try:
                    # Find the prices table (might be named differently)
                    cursor = python_conn.execute(
                        "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '%price%' OR name LIKE '%daily%'"
                    )
                    table_name = cursor.fetchone()
                    if not table_name:
                        # Try common table names
                        for name in ["prices", "daily_prices", "history", "data"]:
                            if self.table_exists(python_conn, name):
                                table_name = (name,)
                                break

                    if not table_name:
                        logger.warning(f"No price table found in {history_file}")
                        continue

                    table_name = table_name[0]
                    python_rows = python_conn.execute(f"SELECT * FROM {table_name}").fetchall()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, table_name)]

                    if not self.dry_run:
                        go_cursor = go_conn.cursor()
                        for row in python_rows:
                            row_dict = dict(zip(python_cols, row))
                            go_cursor.execute("""
                                INSERT OR REPLACE INTO daily_prices (
                                    symbol, date, open, high, low, close, volume, adjusted_close
                                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                            """, (
                                symbol,
                                row_dict.get("date"),
                                row_dict.get("open", 0),
                                row_dict.get("high", 0),
                                row_dict.get("low", 0),
                                row_dict.get("close", 0),
                                row_dict.get("volume"),
                                row_dict.get("adjusted_close") or row_dict.get("close", 0),
                            ))
                        go_conn.commit()
                        total_rows += len(python_rows)
                        logger.info(f"Migrated {len(python_rows)} prices for {symbol}")
                    else:
                        total_rows += len(python_rows)
                        logger.info(f"[DRY RUN] Would migrate {len(python_rows)} prices for {symbol}")

                finally:
                    python_conn.close()

            if not self.dry_run:
                logger.info(f"Total: Migrated {total_rows} price records")
            else:
                logger.info(f"[DRY RUN] Total: Would migrate {total_rows} price records")

        finally:
            go_conn.close()

    def migrate_config(self):
        """Migrate settings and allocation targets to config.db"""
        logger.info("=" * 60)
        logger.info("Migrating Config Database")
        logger.info("=" * 60)

        python_config = self.python_data_dir / "config.db"
        go_config = self.go_data_dir / "config.db"

        if not python_config.exists():
            logger.warning(f"Python config.db not found: {python_config}")
            return

        # Initialize Go config schema
        if not self.dry_run:
            self._init_config_schema(go_config)

        python_conn = sqlite3.connect(python_config)
        go_conn = sqlite3.connect(go_config)

        try:
            # Migrate settings
            if self.table_exists(python_conn, "settings"):
                python_rows = python_conn.execute("SELECT * FROM settings").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, "settings")]

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        go_cursor.execute("""
                            INSERT OR REPLACE INTO settings (key, value, description, updated_at)
                            VALUES (?, ?, ?, ?)
                        """, (
                            row_dict.get("key"),
                            row_dict.get("value"),
                            row_dict.get("description"),
                            row_dict.get("updated_at", datetime.now().isoformat()),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} settings")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} settings")

            # Migrate allocation_targets
            if self.table_exists(python_conn, "allocation_targets"):
                python_rows = python_conn.execute("SELECT * FROM allocation_targets").fetchall()
                if not self.dry_run:
                    go_cursor = go_conn.cursor()
                    python_cols = [col[0] for col in self.get_table_schema(python_conn, "allocation_targets")]
                    now = datetime.now().isoformat()

                    for row in python_rows:
                        row_dict = dict(zip(python_cols, row))
                        # Go schema: id, type, name, target_pct, created_at, updated_at
                        # Map Python columns to Go schema
                        target_type = row_dict.get("type") or row_dict.get("key", "geography")
                        target_name = row_dict.get("name") or row_dict.get("country") or row_dict.get("group") or ""
                        target_pct = row_dict.get("target_pct") or row_dict.get("value") or row_dict.get("target", 0)

                        go_cursor.execute("""
                            INSERT OR REPLACE INTO allocation_targets (
                                type, name, target_pct, created_at, updated_at
                            ) VALUES (?, ?, ?, ?, ?)
                        """, (
                            target_type,
                            target_name,
                            target_pct,
                            row_dict.get("created_at", now),
                            row_dict.get("updated_at", now),
                        ))
                    go_conn.commit()
                    logger.info(f"Migrated {len(python_rows)} allocation targets")
                else:
                    logger.info(f"[DRY RUN] Would migrate {len(python_rows)} allocation targets")

        finally:
            python_conn.close()
            go_conn.close()

    def migrate_cache(self):
        """Migrate recommendations to cache.db"""
        logger.info("=" * 60)
        logger.info("Migrating Cache Database")
        logger.info("=" * 60)

        python_config = self.python_data_dir / "config.db"
        python_recommendations = self.python_data_dir / "recommendations.db"
        go_cache = self.go_data_dir / "cache.db"

        # Initialize Go cache schema
        if not self.dry_run:
            self._init_cache_schema(go_cache)

        go_conn = sqlite3.connect(go_cache)

        try:
            # Try config.db first
            if python_config.exists():
                python_conn = sqlite3.connect(python_config)
                if self.table_exists(python_conn, "recommendations"):
                    python_rows = python_conn.execute("SELECT * FROM recommendations").fetchall()
                    if not self.dry_run:
                        go_cursor = go_conn.cursor()
                        python_cols = [col[0] for col in self.get_table_schema(python_conn, "recommendations")]

                        for row in python_rows:
                            row_dict = dict(zip(python_cols, row))
                            # Default action if missing
                            action = row_dict.get("action") or row_dict.get("recommendation") or "buy"
                            go_cursor.execute("""
                                INSERT OR REPLACE INTO recommendations (
                                    id, symbol, action, priority, score, reason,
                                    created_at, expires_at
                                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                            """, (
                                row_dict.get("id"),
                                row_dict.get("symbol") or "",
                                action,
                                row_dict.get("priority", 0),
                                row_dict.get("score"),
                                row_dict.get("reason"),
                                row_dict.get("created_at", datetime.now().isoformat()),
                                row_dict.get("expires_at"),
                            ))
                        go_conn.commit()
                        logger.info(f"Migrated {len(python_rows)} recommendations from config.db")
                    else:
                        logger.info(f"[DRY RUN] Would migrate {len(python_rows)} recommendations from config.db")
                python_conn.close()

            # Try standalone recommendations.db
            if python_recommendations.exists():
                python_conn = sqlite3.connect(python_recommendations)
                if self.table_exists(python_conn, "recommendations"):
                    python_rows = python_conn.execute("SELECT * FROM recommendations").fetchall()
                    if not self.dry_run:
                        go_cursor = go_conn.cursor()
                        python_cols = [col[0] for col in self.get_table_schema(python_conn, "recommendations")]

                        for row in python_rows:
                            row_dict = dict(zip(python_cols, row))
                            # Default action if missing
                            action = row_dict.get("action") or row_dict.get("recommendation") or "buy"
                            go_cursor.execute("""
                                INSERT OR REPLACE INTO recommendations (
                                    id, symbol, action, priority, score, reason,
                                    created_at, expires_at
                                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                            """, (
                                row_dict.get("id"),
                                row_dict.get("symbol") or "",
                                action,
                                row_dict.get("priority", 0),
                                row_dict.get("score"),
                                row_dict.get("reason"),
                                row_dict.get("created_at", datetime.now().isoformat()),
                                row_dict.get("expires_at"),
                            ))
                        go_conn.commit()
                        logger.info(f"Migrated {len(python_rows)} recommendations from recommendations.db")
                    else:
                        logger.info(f"[DRY RUN] Would migrate {len(python_rows)} recommendations from recommendations.db")
                python_conn.close()

        finally:
            go_conn.close()

    def _init_universe_schema(self, db_path: Path):
        """Initialize universe.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS securities (
                    symbol TEXT PRIMARY KEY,
                    yahoo_symbol TEXT,
                    isin TEXT,
                    name TEXT NOT NULL,
                    product_type TEXT,
                    industry TEXT,
                    country TEXT,
                    fullExchangeName TEXT,
                    priority_multiplier REAL DEFAULT 1.0,
                    min_lot INTEGER DEFAULT 1,
                    active INTEGER DEFAULT 1,
                    allow_buy INTEGER DEFAULT 1,
                    allow_sell INTEGER DEFAULT 1,
                    currency TEXT,
                    last_synced TEXT,
                    min_portfolio_target REAL,
                    max_portfolio_target REAL,
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL,
                    bucket_id TEXT DEFAULT 'core'
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_securities_active ON securities(active);
                CREATE INDEX IF NOT EXISTS idx_securities_bucket ON securities(bucket_id);
                CREATE INDEX IF NOT EXISTS idx_securities_country ON securities(country);
                CREATE INDEX IF NOT EXISTS idx_securities_industry ON securities(industry);
                CREATE INDEX IF NOT EXISTS idx_securities_isin ON securities(isin);

                CREATE TABLE IF NOT EXISTS country_groups (
                    group_name TEXT NOT NULL,
                    country_name TEXT NOT NULL,
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL,
                    PRIMARY KEY (group_name, country_name)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_country_groups_group ON country_groups(group_name);

                CREATE TABLE IF NOT EXISTS industry_groups (
                    group_name TEXT NOT NULL,
                    industry_name TEXT NOT NULL,
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL,
                    PRIMARY KEY (group_name, industry_name)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_industry_groups_group ON industry_groups(group_name);
            """)
            conn.commit()
        finally:
            conn.close()

    def _init_portfolio_schema(self, db_path: Path):
        """Initialize portfolio.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS positions (
                    symbol TEXT PRIMARY KEY,
                    quantity REAL NOT NULL,
                    avg_price REAL NOT NULL,
                    current_price REAL,
                    currency TEXT,
                    currency_rate REAL DEFAULT 1.0,
                    market_value_eur REAL,
                    cost_basis_eur REAL,
                    unrealized_pnl REAL,
                    unrealized_pnl_pct REAL,
                    last_updated TEXT,
                    first_bought TEXT,
                    last_sold TEXT,
                    isin TEXT,
                    bucket_id TEXT DEFAULT 'core'
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_positions_bucket ON positions(bucket_id);
                CREATE INDEX IF NOT EXISTS idx_positions_value ON positions(market_value_eur DESC);

                CREATE TABLE IF NOT EXISTS scores (
                    symbol TEXT PRIMARY KEY,
                    total_score REAL NOT NULL,
                    quality_score REAL,
                    opportunity_score REAL,
                    analyst_score REAL,
                    allocation_fit_score REAL,
                    volatility REAL,
                    cagr_score REAL,
                    consistency_score REAL,
                    history_years INTEGER,
                    technical_score REAL,
                    fundamental_score REAL,
                    last_updated TEXT NOT NULL
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_scores_total ON scores(total_score DESC);
                CREATE INDEX IF NOT EXISTS idx_scores_updated ON scores(last_updated);

                CREATE TABLE IF NOT EXISTS calculated_metrics (
                    symbol TEXT NOT NULL,
                    metric_name TEXT NOT NULL,
                    metric_value REAL NOT NULL,
                    calculated_at TEXT NOT NULL,
                    PRIMARY KEY (symbol, metric_name)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_metrics_symbol ON calculated_metrics(symbol);
                CREATE INDEX IF NOT EXISTS idx_metrics_calculated ON calculated_metrics(calculated_at);

                CREATE TABLE IF NOT EXISTS portfolio_snapshots (
                    snapshot_date TEXT PRIMARY KEY,
                    total_value REAL NOT NULL,
                    cash_balance REAL NOT NULL,
                    invested_value REAL NOT NULL,
                    total_pnl REAL,
                    total_pnl_pct REAL,
                    position_count INTEGER NOT NULL,
                    bucket_id TEXT DEFAULT 'core',
                    snapshot_json TEXT,
                    created_at TEXT NOT NULL
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_snapshots_date ON portfolio_snapshots(snapshot_date DESC);
                CREATE INDEX IF NOT EXISTS idx_snapshots_bucket ON portfolio_snapshots(bucket_id);
            """)
            conn.commit()
        finally:
            conn.close()

    def _init_ledger_schema(self, db_path: Path):
        """Initialize ledger.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            # Drop existing tables if they have old schemas (we'll recreate them)
            # This ensures clean schema migration
            logger.info("Initializing ledger schema (dropping old tables if needed)...")
            conn.execute("DROP TABLE IF EXISTS drip_tracking")
            conn.execute("DROP TABLE IF EXISTS dividend_history")
            conn.execute("DROP TABLE IF EXISTS cash_flows")
            conn.execute("DROP TABLE IF EXISTS trades")
            conn.commit()

            conn.executescript("""
                CREATE TABLE IF NOT EXISTS trades (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    symbol TEXT NOT NULL,
                    isin TEXT,
                    side TEXT NOT NULL CHECK (side IN ('buy', 'sell')),
                    quantity REAL NOT NULL CHECK (quantity > 0),
                    price REAL NOT NULL CHECK (price > 0),
                    executed_at TEXT NOT NULL,
                    order_id TEXT,
                    currency TEXT NOT NULL,
                    value_eur REAL NOT NULL,
                    source TEXT DEFAULT 'manual',
                    bucket_id TEXT DEFAULT 'core',
                    mode TEXT DEFAULT 'normal',
                    created_at TEXT NOT NULL
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
                CREATE INDEX IF NOT EXISTS idx_trades_executed ON trades(executed_at DESC);
                CREATE INDEX IF NOT EXISTS idx_trades_bucket ON trades(bucket_id);

                CREATE TABLE IF NOT EXISTS cash_flows (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    flow_type TEXT NOT NULL CHECK (flow_type IN ('deposit', 'withdrawal', 'fee', 'dividend', 'interest')),
                    amount REAL NOT NULL,
                    currency TEXT NOT NULL,
                    amount_eur REAL NOT NULL,
                    description TEXT,
                    executed_at TEXT NOT NULL,
                    bucket_id TEXT DEFAULT 'core',
                    created_at TEXT NOT NULL
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_cashflows_type ON cash_flows(flow_type);
                CREATE INDEX IF NOT EXISTS idx_cashflows_executed ON cash_flows(executed_at DESC);
                CREATE INDEX IF NOT EXISTS idx_cashflows_bucket ON cash_flows(bucket_id);

                CREATE TABLE IF NOT EXISTS dividend_history (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    symbol TEXT NOT NULL,
                    isin TEXT,
                    payment_date TEXT NOT NULL,
                    ex_date TEXT,
                    amount_per_share REAL NOT NULL,
                    shares_held REAL NOT NULL,
                    total_amount REAL NOT NULL,
                    currency TEXT NOT NULL,
                    total_amount_eur REAL NOT NULL,
                    drip_enabled INTEGER DEFAULT 0,
                    reinvested_shares REAL,
                    bucket_id TEXT DEFAULT 'core',
                    created_at TEXT NOT NULL
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_dividends_symbol ON dividend_history(symbol);
                CREATE INDEX IF NOT EXISTS idx_dividends_payment_date ON dividend_history(payment_date DESC);
                CREATE INDEX IF NOT EXISTS idx_dividends_bucket ON dividend_history(bucket_id);

                CREATE TABLE IF NOT EXISTS drip_tracking (
                    symbol TEXT PRIMARY KEY,
                    drip_enabled INTEGER DEFAULT 0,
                    total_dividends_received REAL DEFAULT 0,
                    total_shares_reinvested REAL DEFAULT 0,
                    last_dividend_date TEXT,
                    updated_at TEXT NOT NULL
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_drip_enabled ON drip_tracking(drip_enabled);
            """)
            conn.commit()
        finally:
            conn.close()

    def _init_satellites_schema(self, db_path: Path):
        """Initialize satellites.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS buckets (
                    id TEXT PRIMARY KEY,
                    name TEXT NOT NULL,
                    type TEXT NOT NULL,
                    notes TEXT,
                    target_pct REAL,
                    min_pct REAL,
                    max_pct REAL,
                    consecutive_losses INTEGER DEFAULT 0,
                    max_consecutive_losses INTEGER DEFAULT 5,
                    high_water_mark REAL DEFAULT 0,
                    high_water_mark_date TEXT,
                    loss_streak_paused_at TEXT,
                    status TEXT DEFAULT 'active',
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL
                );

                CREATE INDEX IF NOT EXISTS idx_buckets_type ON buckets(type);
                CREATE INDEX IF NOT EXISTS idx_buckets_status ON buckets(status);

                CREATE TABLE IF NOT EXISTS satellite_settings (
                    satellite_id TEXT PRIMARY KEY,
                    preset TEXT,
                    risk_appetite REAL DEFAULT 0.5,
                    hold_duration REAL DEFAULT 0.5,
                    entry_style REAL DEFAULT 0.5,
                    position_spread REAL DEFAULT 0.5,
                    profit_taking REAL DEFAULT 0.5,
                    trailing_stops INTEGER DEFAULT 0,
                    follow_regime INTEGER DEFAULT 0,
                    auto_harvest INTEGER DEFAULT 0,
                    pause_high_volatility INTEGER DEFAULT 0,
                    dividend_handling TEXT DEFAULT 'reinvest_same',
                    risk_free_rate REAL DEFAULT 0.035,
                    sortino_mar REAL DEFAULT 0.05,
                    evaluation_period_days INTEGER DEFAULT 90,
                    volatility_window INTEGER DEFAULT 60,
                    FOREIGN KEY (satellite_id) REFERENCES buckets(id) ON DELETE CASCADE
                );

                CREATE TABLE IF NOT EXISTS bucket_balances (
                    bucket_id TEXT NOT NULL,
                    currency TEXT NOT NULL,
                    balance REAL NOT NULL DEFAULT 0,
                    last_updated TEXT NOT NULL,
                    PRIMARY KEY (bucket_id, currency),
                    FOREIGN KEY (bucket_id) REFERENCES buckets(id) ON DELETE CASCADE
                );

                CREATE INDEX IF NOT EXISTS idx_bucket_balances_bucket ON bucket_balances(bucket_id);

                CREATE TABLE IF NOT EXISTS bucket_transactions (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    bucket_id TEXT NOT NULL,
                    type TEXT NOT NULL,
                    amount REAL NOT NULL,
                    currency TEXT NOT NULL,
                    description TEXT,
                    created_at TEXT NOT NULL,
                    FOREIGN KEY (bucket_id) REFERENCES buckets(id) ON DELETE CASCADE
                );

                CREATE INDEX IF NOT EXISTS idx_bucket_transactions_bucket ON bucket_transactions(bucket_id);
                CREATE INDEX IF NOT EXISTS idx_bucket_transactions_type ON bucket_transactions(type);
                CREATE INDEX IF NOT EXISTS idx_bucket_transactions_created ON bucket_transactions(created_at);

                CREATE TABLE IF NOT EXISTS allocation_settings (
                    key TEXT PRIMARY KEY,
                    value REAL NOT NULL,
                    description TEXT
                );

                CREATE TABLE IF NOT EXISTS satellite_regime_performance (
                    satellite_id TEXT NOT NULL,
                    regime TEXT NOT NULL,
                    period_start TEXT NOT NULL,
                    period_end TEXT NOT NULL,
                    return_pct REAL,
                    trades_count INTEGER,
                    win_rate REAL,
                    PRIMARY KEY (satellite_id, regime, period_start),
                    FOREIGN KEY (satellite_id) REFERENCES buckets(id) ON DELETE CASCADE
                );

                CREATE INDEX IF NOT EXISTS idx_regime_performance_satellite ON satellite_regime_performance(satellite_id);
            """)
            conn.commit()
        finally:
            conn.close()

    def _init_agents_schema(self, db_path: Path):
        """Initialize agents.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS sequences (
                    sequence_hash TEXT NOT NULL,
                    portfolio_hash TEXT NOT NULL,
                    priority REAL NOT NULL,
                    sequence_json TEXT NOT NULL,
                    depth INTEGER NOT NULL,
                    pattern_type TEXT,
                    completed INTEGER DEFAULT 0,
                    evaluated_at TEXT,
                    created_at TEXT NOT NULL,
                    PRIMARY KEY (sequence_hash, portfolio_hash)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_sequences_portfolio ON sequences(portfolio_hash);
                CREATE INDEX IF NOT EXISTS idx_sequences_priority ON sequences(portfolio_hash, priority DESC, completed);
                CREATE INDEX IF NOT EXISTS idx_sequences_completed ON sequences(portfolio_hash, completed);

                CREATE TABLE IF NOT EXISTS evaluations (
                    sequence_hash TEXT NOT NULL,
                    portfolio_hash TEXT NOT NULL,
                    end_score REAL NOT NULL,
                    breakdown_json TEXT NOT NULL,
                    end_cash REAL NOT NULL,
                    end_context_positions_json TEXT NOT NULL,
                    div_score REAL NOT NULL,
                    total_value REAL NOT NULL,
                    evaluated_at TEXT NOT NULL,
                    PRIMARY KEY (sequence_hash, portfolio_hash)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_evaluations_portfolio ON evaluations(portfolio_hash);
                CREATE INDEX IF NOT EXISTS idx_evaluations_score ON evaluations(portfolio_hash, end_score DESC);

                CREATE TABLE IF NOT EXISTS best_result (
                    portfolio_hash TEXT PRIMARY KEY,
                    best_sequence_hash TEXT NOT NULL,
                    best_score REAL NOT NULL,
                    updated_at TEXT NOT NULL
                ) STRICT;

                CREATE TABLE IF NOT EXISTS agent_configs (
                    id TEXT PRIMARY KEY,
                    name TEXT NOT NULL,
                    toml_config TEXT NOT NULL,
                    bucket_id TEXT,
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_agent_configs_bucket ON agent_configs(bucket_id);
                CREATE INDEX IF NOT EXISTS idx_agent_configs_name ON agent_configs(name);

                CREATE TABLE IF NOT EXISTS config_history (
                    id TEXT PRIMARY KEY,
                    agent_config_id TEXT NOT NULL,
                    name TEXT NOT NULL,
                    toml_config TEXT NOT NULL,
                    saved_at TEXT NOT NULL,
                    performance_score REAL,
                    FOREIGN KEY (agent_config_id) REFERENCES agent_configs(id) ON DELETE CASCADE
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_config_history_agent ON config_history(agent_config_id);
                CREATE INDEX IF NOT EXISTS idx_config_history_saved ON config_history(saved_at DESC);
                CREATE INDEX IF NOT EXISTS idx_config_history_performance ON config_history(agent_config_id, performance_score DESC);
            """)
            conn.commit()
        finally:
            conn.close()

    def _init_history_schema(self, db_path: Path):
        """Initialize history.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS daily_prices (
                    symbol TEXT NOT NULL,
                    date TEXT NOT NULL,
                    open REAL NOT NULL,
                    high REAL NOT NULL,
                    low REAL NOT NULL,
                    close REAL NOT NULL,
                    volume INTEGER,
                    adjusted_close REAL,
                    PRIMARY KEY (symbol, date)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_prices_symbol ON daily_prices(symbol);
                CREATE INDEX IF NOT EXISTS idx_prices_date ON daily_prices(date DESC);
                CREATE INDEX IF NOT EXISTS idx_prices_symbol_date ON daily_prices(symbol, date DESC);

                CREATE TABLE IF NOT EXISTS exchange_rates (
                    from_currency TEXT NOT NULL,
                    to_currency TEXT NOT NULL,
                    date TEXT NOT NULL,
                    rate REAL NOT NULL,
                    PRIMARY KEY (from_currency, to_currency, date)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_rates_pair ON exchange_rates(from_currency, to_currency);
                CREATE INDEX IF NOT EXISTS idx_rates_date ON exchange_rates(date DESC);

                CREATE TABLE IF NOT EXISTS symbol_removals (
                    symbol TEXT PRIMARY KEY,
                    removed_at INTEGER NOT NULL,
                    grace_period_days INTEGER DEFAULT 30,
                    row_count INTEGER,
                    marked_by TEXT
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_removals_date ON symbol_removals(removed_at);

                CREATE TABLE IF NOT EXISTS cleanup_log (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    symbol TEXT NOT NULL,
                    deleted_at INTEGER NOT NULL,
                    row_count INTEGER,
                    cleanup_reason TEXT,
                    size_freed_bytes INTEGER
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_cleanup_symbol ON cleanup_log(symbol);
                CREATE INDEX IF NOT EXISTS idx_cleanup_date ON cleanup_log(deleted_at DESC);
            """)
            conn.commit()
        finally:
            conn.close()

    def _init_config_schema(self, db_path: Path):
        """Initialize config.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS settings (
                    key TEXT PRIMARY KEY,
                    value TEXT NOT NULL,
                    description TEXT,
                    updated_at TEXT NOT NULL
                ) STRICT;

                CREATE TABLE IF NOT EXISTS allocation_targets (
                    id INTEGER PRIMARY KEY,
                    type TEXT NOT NULL,
                    name TEXT NOT NULL,
                    target_pct REAL NOT NULL,
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL,
                    UNIQUE(type, name)
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_allocation_type ON allocation_targets(type);
            """)
            conn.commit()
        finally:
            conn.close()

    def _init_cache_schema(self, db_path: Path):
        """Initialize cache.db schema."""
        conn = sqlite3.connect(db_path)
        try:
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS recommendations (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    symbol TEXT NOT NULL,
                    action TEXT NOT NULL,
                    priority REAL NOT NULL,
                    score REAL,
                    reason TEXT,
                    created_at TEXT NOT NULL,
                    expires_at TEXT
                ) STRICT;

                CREATE INDEX IF NOT EXISTS idx_recommendations_symbol ON recommendations(symbol);
                CREATE INDEX IF NOT EXISTS idx_recommendations_priority ON recommendations(priority DESC);
                CREATE INDEX IF NOT EXISTS idx_recommendations_created ON recommendations(created_at DESC);
            """)
            conn.commit()
        finally:
            conn.close()

    def run_migration(self):
        """Run the complete migration process."""
        logger.info("=" * 80)
        logger.info("Starting Python to Go Database Migration")
        logger.info("=" * 80)
        logger.info(f"Python data directory: {self.python_data_dir}")
        logger.info(f"Go data directory: {self.go_data_dir}")
        logger.info(f"Dry run: {self.dry_run}")
        logger.info("=" * 80)

        if not self.python_data_dir.exists():
            logger.error(f"Python data directory does not exist: {self.python_data_dir}")
            return False

        # Backup Python databases
        logger.info("\nBacking up Python databases...")
        python_dbs = [
            "config.db", "state.db", "ledger.db", "calculations.db",
            "snapshots.db", "dividends.db", "satellites.db", "planner.db",
            "recommendations.db", "rates.db"
        ]

        for db_name in python_dbs:
            db_path = self.python_data_dir / db_name
            if db_path.exists():
                self.backup_database(db_path)

        # Run migrations
        logger.info("\nRunning migrations...")
        self.migrate_universe()
        self.migrate_portfolio()
        self.migrate_ledger()
        self.migrate_satellites()
        self.migrate_agents()
        self.migrate_history()
        self.migrate_config()
        self.migrate_cache()

        logger.info("\n" + "=" * 80)
        logger.info("Migration completed!")
        logger.info("=" * 80)

        if not self.dry_run:
            logger.info(f"Backups saved to: {self.backup_dir}")

        return True


def main():
    parser = argparse.ArgumentParser(description="Migrate Python app databases to Go app databases")
    parser.add_argument(
        "--python-data-dir",
        type=str,
        default="data",
        help="Path to Python app data directory (default: data)"
    )
    parser.add_argument(
        "--go-data-dir",
        type=str,
        default="app/data",
        help="Path to Go app data directory (default: app/data)"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simulate migration without making changes"
    )

    args = parser.parse_args()

    python_data_dir = Path(args.python_data_dir).expanduser().resolve()
    go_data_dir = Path(args.go_data_dir).expanduser().resolve()

    migrator = DatabaseMigrator(python_data_dir, go_data_dir, args.dry_run)
    success = migrator.run_migration()

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
