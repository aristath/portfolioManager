"""Monthly rebalance job."""

import logging
from datetime import datetime

import aiosqlite

from app.config import settings

logger = logging.getLogger(__name__)


async def execute_monthly_rebalance():
    """
    Execute monthly portfolio rebalance.

    This job:
    1. Syncs portfolio from Tradernet
    2. Refreshes stock scores
    3. Calculates optimal trades for the monthly deposit
    4. Executes trades via Tradernet
    """
    logger.info("Starting monthly rebalance")

    try:
        from app.jobs.daily_sync import sync_portfolio
        from app.services.scorer import score_all_stocks
        from app.services.allocator import calculate_rebalance_trades, execute_trades

        # Step 1: Sync portfolio
        logger.info("Step 1: Syncing portfolio...")
        await sync_portfolio()

        async with aiosqlite.connect(settings.database_path) as db:
            # Step 2: Refresh scores
            logger.info("Step 2: Refreshing stock scores...")
            scores = await score_all_stocks(db)
            logger.info(f"Scored {len(scores)} stocks")

            # Step 3: Calculate rebalance trades
            logger.info("Step 3: Calculating rebalance trades...")
            trades = await calculate_rebalance_trades(db, settings.monthly_deposit)

            if not trades:
                logger.info("No rebalance trades needed")
                return

            logger.info(f"Generated {len(trades)} trade recommendations:")
            for trade in trades:
                logger.info(
                    f"  {trade.side} {trade.quantity} {trade.symbol} "
                    f"@ {trade.estimated_price:.2f} ({trade.reason})"
                )

            # Step 4: Execute trades
            logger.info("Step 4: Executing trades...")
            results = await execute_trades(db, trades)

            successful = sum(1 for r in results if r["status"] == "success")
            failed = sum(1 for r in results if r["status"] != "success")

            logger.info(
                f"Monthly rebalance complete: {successful} successful, {failed} failed"
            )

            # Create snapshot after rebalance
            await sync_portfolio()

    except Exception as e:
        logger.error(f"Monthly rebalance failed: {e}")
        raise
