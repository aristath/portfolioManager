"""Rebalancing application service.

Orchestrates rebalancing operations using domain services and repositories.
"""

import logging
from typing import List

from app.config import settings
from app.domain.repositories import (
    StockRepository,
    PositionRepository,
    AllocationRepository,
    PortfolioRepository,
)
from app.domain.services.priority_calculator import (
    PriorityCalculator,
    PriorityInput,
)
from app.services.allocator import (
    TradeRecommendation,
    StockPriority,
    calculate_position_size,
    get_max_trades,
    parse_industries,
)
from app.services import yahoo

logger = logging.getLogger(__name__)


class RebalancingService:
    """Application service for rebalancing operations."""

    def __init__(
        self,
        stock_repo: StockRepository,
        position_repo: PositionRepository,
        allocation_repo: AllocationRepository,
        portfolio_repo: PortfolioRepository,
    ):
        self._stock_repo = stock_repo
        self._position_repo = position_repo
        self._allocation_repo = allocation_repo
        self._portfolio_repo = portfolio_repo

    async def calculate_rebalance_trades(
        self,
        available_cash: float
    ) -> List[TradeRecommendation]:
        """
        Calculate optimal trades using enhanced multi-factor priority algorithm.

        Strategy:
        1. Only consider stocks with score > min_stock_score
        2. Calculate enhanced priority using PriorityCalculator
        3. Select top N stocks by combined priority
        4. Dynamic position sizing based on conviction/risk
        5. Minimum €400 per trade (min_trade_size)
        6. Maximum 5 trades per cycle (max_trades_per_cycle)
        """
        # Check minimum cash threshold
        if available_cash < settings.min_cash_threshold:
            logger.info(f"Cash €{available_cash:.2f} below minimum €{settings.min_cash_threshold:.2f}")
            return []

        max_trades = get_max_trades(available_cash)
        if max_trades == 0:
            return []

        # Get portfolio summary for weight lookups
        from app.application.services.portfolio_service import PortfolioService
        portfolio_service = PortfolioService(
            self._portfolio_repo,
            self._position_repo,
            self._allocation_repo,
        )
        summary = await portfolio_service.get_portfolio_summary()
        total_value = summary.total_value or 1  # Avoid division by zero

        # Build weight maps for quick lookup (target_pct now stores weights -1 to +1)
        geo_weights = {a.name: a.target_pct for a in summary.geographic_allocations}
        industry_weights = {a.name: a.target_pct for a in summary.industry_allocations}

        # Get scored stocks from universe with volatility, multiplier, and min_lot
        stocks_data = await self._stock_repo.get_with_scores()

        # Calculate priority for each stock
        priority_inputs = []
        stock_metadata = {}  # Store min_lot for later use

        for stock in stocks_data:
            symbol = stock["symbol"]
            name = stock["name"]
            geography = stock["geography"]
            industry = stock.get("industry")
            multiplier = stock.get("priority_multiplier") or 1.0
            min_lot = stock.get("min_lot") or 1
            score = stock.get("total_score") or 0
            volatility = stock.get("volatility")
            position_value = stock.get("position_value") or 0

            # Only consider stocks with score above threshold
            if score < settings.min_stock_score:
                logger.debug(f"Skipping {symbol}: score {score:.2f} < {settings.min_stock_score}")
                continue

            priority_inputs.append(PriorityInput(
                symbol=symbol,
                name=name,
                geography=geography,
                industry=industry,
                stock_score=score,
                volatility=volatility,
                multiplier=multiplier,
                position_value=position_value,
                total_portfolio_value=total_value,
            ))

            stock_metadata[symbol] = {
                "min_lot": min_lot,
                "name": name,
                "geography": geography,
                "industry": industry,
            }

        if not priority_inputs:
            logger.warning("No stocks qualify for purchase (all scores below threshold)")
            return []

        # Calculate priorities using domain service
        priority_results = PriorityCalculator.calculate_priorities(
            priority_inputs,
            geo_weights,
            industry_weights,
        )

        logger.info(f"Found {len(priority_results)} qualified stocks (score >= {settings.min_stock_score})")

        # Select top N candidates
        selected = priority_results[:max_trades]

        # Calculate base trade size per stock
        base_trade_size = available_cash / len(selected)

        # Get current prices and generate recommendations with dynamic sizing
        recommendations = []
        remaining_cash = available_cash

        for result in selected:
            if remaining_cash < settings.min_trade_size:
                break

            metadata = stock_metadata[result.symbol]

            # Get current price from Yahoo Finance
            # Note: This could be moved to a price service in the future
            # We need yahoo_symbol for proper price lookup - get it from stock data
            stock_data = next((s for s in stocks_data if s["symbol"] == result.symbol), None)
            yahoo_symbol = stock_data.get("yahoo_symbol") if stock_data else None
            price = yahoo.get_current_price(result.symbol, yahoo_symbol)
            if not price or price <= 0:
                logger.warning(f"Could not get price for {result.symbol}, skipping")
                continue

            # Create StockPriority for position sizing calculation
            candidate = StockPriority(
                symbol=result.symbol,
                name=result.name,
                geography=result.geography,
                industry=result.industry,
                stock_score=result.stock_score,
                volatility=result.volatility,
                multiplier=result.multiplier,
                min_lot=metadata["min_lot"],
                geo_need=result.geo_need,
                industry_need=result.industry_need,
                combined_priority=result.combined_priority,
            )

            # Dynamic position sizing based on conviction and risk
            dynamic_size = calculate_position_size(
                candidate,
                base_trade_size,
                settings.min_trade_size
            )
            invest_amount = min(dynamic_size, remaining_cash)
            if invest_amount < settings.min_trade_size:
                continue

            # Calculate quantity with minimum lot size rounding
            min_lot = metadata["min_lot"]
            lot_cost = min_lot * price

            # Check if we can afford at least one lot
            if lot_cost > invest_amount:
                logger.debug(
                    f"Skipping {result.symbol}: min lot {min_lot} @ €{price:.2f} = "
                    f"€{lot_cost:.2f} > available €{invest_amount:.2f}"
                )
                continue

            # Calculate how many lots we can buy (rounding down to whole lots)
            num_lots = int(invest_amount / lot_cost)
            qty = num_lots * min_lot

            if qty <= 0:
                continue

            actual_value = qty * price

            # Build reason string with more detail
            reason_parts = []
            if result.geo_need > 0.6:
                reason_parts.append(f"{result.geography} prioritized")
            elif result.geo_need < 0.4:
                reason_parts.append(f"{result.geography} neutral")
            if result.industry_need > 0.6:
                industries = parse_industries(result.industry)
                if industries:
                    reason_parts.append(f"{industries[0]} prioritized")
            reason_parts.append(f"score: {result.stock_score:.2f}")
            if result.multiplier != 1.0:
                reason_parts.append(f"mult: {result.multiplier:.1f}x")
            reason = ", ".join(reason_parts)

            recommendations.append(TradeRecommendation(
                symbol=result.symbol,
                name=result.name,
                side="BUY",
                quantity=qty,
                estimated_price=round(price, 2),
                estimated_value=round(actual_value, 2),
                reason=reason,
            ))

            remaining_cash -= actual_value

        total_invested = available_cash - remaining_cash
        logger.info(
            f"Generated {len(recommendations)} trade recommendations, "
            f"total value: €{total_invested:.2f} from €{available_cash:.2f}"
        )

        return recommendations
