"""APScheduler setup for background jobs."""

import logging
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger
from apscheduler.triggers.interval import IntervalTrigger

from app.config import settings

logger = logging.getLogger(__name__)

# Global scheduler instance
scheduler: AsyncIOScheduler = None


def init_scheduler() -> AsyncIOScheduler:
    """Initialize the APScheduler."""
    global scheduler

    scheduler = AsyncIOScheduler()

    # Import jobs here to avoid circular imports
    from app.jobs.daily_sync import sync_portfolio, sync_prices
    from app.jobs.cash_rebalance import check_and_rebalance

    # Daily portfolio sync (at configured hour)
    scheduler.add_job(
        sync_portfolio,
        CronTrigger(hour=settings.daily_sync_hour, minute=0),
        id="daily_portfolio_sync",
        name="Daily Portfolio Sync",
        replace_existing=True,
    )

    # Daily price sync (every 4 hours during market hours)
    scheduler.add_job(
        sync_prices,
        CronTrigger(hour="9,13,17,21", minute=0),
        id="price_sync",
        name="Price Sync",
        replace_existing=True,
    )

    # Cash-based rebalance check (every N minutes)
    scheduler.add_job(
        check_and_rebalance,
        IntervalTrigger(minutes=settings.cash_check_interval_minutes),
        id="cash_rebalance_check",
        name="Cash Rebalance Check",
        replace_existing=True,
    )

    logger.info("Scheduler initialized with jobs")
    return scheduler


def start_scheduler():
    """Start the scheduler."""
    global scheduler
    if scheduler and not scheduler.running:
        scheduler.start()
        logger.info("Scheduler started")


def stop_scheduler():
    """Stop the scheduler."""
    global scheduler
    if scheduler and scheduler.running:
        scheduler.shutdown()
        logger.info("Scheduler stopped")


def get_scheduler() -> AsyncIOScheduler:
    """Get the scheduler instance."""
    global scheduler
    if scheduler is None:
        scheduler = init_scheduler()
    return scheduler
