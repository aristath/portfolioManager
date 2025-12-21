"""Cash-based rebalance job.

Checks cash balance periodically and executes trades when sufficient funds available.
Replaces the fixed monthly deposit schedule with dynamic cash-triggered rebalancing.
"""

import logging
from datetime import datetime

import aiosqlite

from app.config import settings
from app.services.tradernet import get_tradernet_client

logger = logging.getLogger(__name__)


async def check_and_rebalance():
    """
    Check cash balance and execute rebalancing if threshold is met.

    This job runs every 15 minutes and:
    1. Gets current cash balance from Tradernet
    2. If cash >= min_cash_threshold (€400), calculates and executes trades
    3. Logs the result

    Trade sizing:
    - Minimum €400 per trade (keeps commission at 0.5%)
    - Maximum 5 trades per cycle
    - Formula: max_trades = min(5, floor(cash / 400))
    """
    logger.info("Checking cash balance for potential rebalance...")

    try:
        # Get cash balance from Tradernet
        client = get_tradernet_client()
        if not client.is_connected:
            if not client.connect():
                logger.warning("Cannot connect to Tradernet, skipping cash check")
                return

        cash_balance = client.get_total_cash_eur()
        logger.info(f"Current cash balance: €{cash_balance:.2f}")

        # Check if we have enough cash to trade
        if cash_balance < settings.min_cash_threshold:
            logger.info(
                f"Cash €{cash_balance:.2f} below threshold €{settings.min_cash_threshold:.2f}, "
                "no rebalance needed"
            )
            return

        # We have enough cash - proceed with rebalancing
        logger.info(f"Cash €{cash_balance:.2f} >= threshold, initiating rebalance")

        from app.jobs.daily_sync import sync_portfolio
        from app.services.scorer import score_all_stocks
        from app.services.allocator import calculate_rebalance_trades, execute_trades

        # Step 1: Sync portfolio to get latest positions
        logger.info("Step 1: Syncing portfolio...")
        await sync_portfolio()

        async with aiosqlite.connect(settings.database_path) as db:
            # Step 2: Refresh scores
            logger.info("Step 2: Refreshing stock scores...")
            scores = await score_all_stocks(db)
            logger.info(f"Scored {len(scores)} stocks")

            # Step 3: Calculate rebalance trades based on available cash
            logger.info("Step 3: Calculating rebalance trades...")
            trades = await calculate_rebalance_trades(db, cash_balance)

            if not trades:
                logger.info("No rebalance trades recommended")
                return

            logger.info(f"Generated {len(trades)} trade recommendations:")
            for trade in trades:
                logger.info(
                    f"  {trade.side} {trade.quantity} {trade.symbol} "
                    f"@ €{trade.estimated_price:.2f} = €{trade.estimated_value:.2f} "
                    f"({trade.reason})"
                )

            # Step 4: Execute trades
            logger.info("Step 4: Executing trades...")
            results = await execute_trades(db, trades)

            successful = sum(1 for r in results if r["status"] == "success")
            failed = sum(1 for r in results if r["status"] != "success")

            logger.info(
                f"Rebalance complete: {successful} successful, {failed} failed"
            )

            # Create snapshot after rebalance
            await sync_portfolio()

            # Log summary
            total_invested = sum(t.estimated_value for t in trades)
            logger.info(
                f"Cash rebalance finished at {datetime.now().isoformat()}: "
                f"invested €{total_invested:.2f} from €{cash_balance:.2f} available"
            )

    except Exception as e:
        logger.error(f"Cash rebalance check failed: {e}")
        raise
