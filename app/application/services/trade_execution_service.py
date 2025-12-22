"""Trade execution application service.

Orchestrates trade execution via Tradernet and records trades.
"""

import logging
from typing import List
from datetime import datetime

from app.domain.repositories import TradeRepository, Trade
from app.services.allocator import TradeRecommendation
from app.services.tradernet import get_tradernet_client

logger = logging.getLogger(__name__)


class TradeExecutionService:
    """Application service for trade execution."""

    def __init__(self, trade_repo: TradeRepository):
        self._trade_repo = trade_repo

    async def execute_trades(
        self,
        trades: List[TradeRecommendation]
    ) -> List[dict]:
        """
        Execute a list of trade recommendations via Tradernet.

        Args:
            trades: List of trade recommendations to execute

        Returns:
            List of execution results with status for each trade
        """
        client = get_tradernet_client()

        if not client.is_connected:
            if not client.connect():
                raise ConnectionError("Failed to connect to Tradernet")

        results = []

        for trade in trades:
            try:
                result = client.place_order(
                    symbol=trade.symbol,
                    side=trade.side,
                    quantity=trade.quantity,
                )

                if result:
                    # Record trade using repository
                    trade_record = Trade(
                        symbol=trade.symbol,
                        side=trade.side,
                        quantity=trade.quantity,
                        price=result.price or trade.estimated_price,
                        executed_at=datetime.now(),
                        order_id=result.order_id,
                    )
                    await self._trade_repo.create(trade_record)

                    results.append({
                        "symbol": trade.symbol,
                        "status": "success",
                        "order_id": result.order_id,
                    })
                else:
                    results.append({
                        "symbol": trade.symbol,
                        "status": "failed",
                        "error": "Order placement returned None",
                    })

            except Exception as e:
                logger.error(f"Failed to execute trade for {trade.symbol}: {e}")
                results.append({
                    "symbol": trade.symbol,
                    "status": "error",
                    "error": str(e),
                })

        return results
