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
        from app.infrastructure.dependencies import (
            get_stock_repository,
            get_position_repository,
            get_allocation_repository,
            get_portfolio_repository,
            get_trade_repository,
        )
        from app.application.services.rebalancing_service import RebalancingService
        from app.application.services.trade_execution_service import TradeExecutionService

        # Step 1: Sync portfolio
        logger.info("Step 1: Syncing portfolio...")
        await sync_portfolio()

        async with aiosqlite.connect(settings.database_path) as db:
            # Step 2: Refresh scores
            logger.info("Step 2: Refreshing stock scores...")
            scores = await score_all_stocks(db)
            logger.info(f"Scored {len(scores)} stocks")

            # Step 3: Calculate rebalance trades using application services
            logger.info("Step 3: Calculating rebalance trades...")
            stock_repo = get_stock_repository(db)
            position_repo = get_position_repository(db)
            allocation_repo = get_allocation_repository(db)
            portfolio_repo = get_portfolio_repository(db)

            rebalancing_service = RebalancingService(
                stock_repo, position_repo, allocation_repo, portfolio_repo
            )
            trades = await rebalancing_service.calculate_rebalance_trades(settings.monthly_deposit)

            if not trades:
                logger.info("No rebalance trades needed")
                return

            logger.info(f"Generated {len(trades)} trade recommendations:")
            for trade in trades:
                logger.info(
                    f"  {trade.side} {trade.quantity} {trade.symbol} "
                    f"@ {trade.estimated_price:.2f} ({trade.reason})"
                )

            # Step 4: Execute trades using application service
            logger.info("Step 4: Executing trades...")
            trade_repo = get_trade_repository(db)
            trade_execution_service = TradeExecutionService(trade_repo)
            results = await trade_execution_service.execute_trades(trades)

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
