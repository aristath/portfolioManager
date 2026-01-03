"""Per-security performance calculation service for display visualization.

Calculates individual security performance metrics for multi-cluster display.
"""

import logging
from datetime import datetime, timedelta
from typing import Optional

from app.modules.portfolio.database.history_repository import HistoryRepository

logger = logging.getLogger(__name__)


class SecurityPerformanceService:
    """Calculate individual security performance metrics for display."""

    async def calculate_trailing_12mo_cagr(self, symbol: str) -> Optional[float]:
        """Calculate trailing 12-month CAGR for a specific security.

        Args:
            symbol: Security symbol (e.g., "AAPL", "MSFT")

        Returns:
            Trailing 12mo CAGR as decimal (e.g., 0.15 = 15%), or None if insufficient data
        """
        history_repo = HistoryRepository(symbol)

        # Get date range for last 12 months
        end_date = datetime.now()
        start_date = end_date - timedelta(days=365)
        start_date_str = start_date.strftime("%Y-%m-%d")
        end_date_str = end_date.strftime("%Y-%m-%d")

        try:
            # Get daily prices in date range
            prices = await history_repo.get_daily_range(start_date_str, end_date_str)

            if not prices or len(prices) < 2:
                logger.debug(
                    f"Insufficient price data for {symbol} trailing 12mo calculation"
                )
                return None

            # Prices are ordered ASC, so first is start, last is end
            start_price = prices[0].close_price
            end_price = prices[-1].close_price

            if not start_price or start_price <= 0:
                logger.warning(
                    f"Invalid start price for {symbol} trailing 12mo calculation"
                )
                return None

            # Calculate days between first and last price
            start_dt = datetime.strptime(prices[0].date, "%Y-%m-%d")
            end_dt = datetime.strptime(prices[-1].date, "%Y-%m-%d")
            days = (end_dt - start_dt).days

            if days < 30:
                logger.debug(
                    f"Insufficient time period for {symbol} trailing 12mo calculation"
                )
                return None

            # Calculate annualized return
            years = days / 365.0

            if years >= 0.25:
                # Use CAGR formula for periods >= 3 months
                cagr = ((end_price / start_price) ** (1 / years)) - 1
            else:
                # Simple return for very short periods
                cagr = (end_price / start_price) - 1

            logger.debug(
                f"{symbol} trailing 12mo CAGR: {cagr:.4f} "
                f"(from {start_price:.2f} to {end_price:.2f} over {days} days)"
            )
            return cagr

        except Exception as e:
            logger.error(f"Error calculating trailing 12mo CAGR for {symbol}: {e}")
            return None

    async def get_performance_vs_target(
        self, symbol: str, target: float
    ) -> Optional[float]:
        """Get security performance difference vs target.

        Args:
            symbol: Security symbol
            target: Target annual return (e.g., 0.11 = 11%)

        Returns:
            Difference from target (e.g., 0.03 = 3% above target), or None if no data
        """
        cagr = await self.calculate_trailing_12mo_cagr(symbol)
        if cagr is None:
            return None

        difference = cagr - target

        logger.debug(
            f"{symbol} performance vs target: {difference:+.4f} "
            f"(CAGR: {cagr:.4f}, target: {target:.4f})"
        )
        return difference
