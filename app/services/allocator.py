"""Portfolio allocation and rebalancing logic."""

import logging
from dataclasses import dataclass
from typing import Optional

import aiosqlite

from app.config import settings

logger = logging.getLogger(__name__)


@dataclass
class AllocationStatus:
    """Current allocation vs target."""
    category: str  # geography or industry
    name: str  # EU, ASIA, US or Technology, etc.
    target_pct: float
    current_pct: float
    current_value: float
    deviation: float  # current - target (negative = underweight)


@dataclass
class PortfolioSummary:
    """Complete portfolio allocation summary."""
    total_value: float
    cash_balance: float
    geographic_allocations: list[AllocationStatus]
    industry_allocations: list[AllocationStatus]


@dataclass
class TradeRecommendation:
    """Recommended trade for rebalancing."""
    symbol: str
    name: str
    side: str  # BUY or SELL
    quantity: float
    estimated_price: float
    estimated_value: float
    reason: str  # Why this trade is recommended


async def get_portfolio_summary(db: aiosqlite.Connection) -> PortfolioSummary:
    """
    Calculate current portfolio allocation vs targets.

    Returns complete summary with geographic and industry breakdowns.
    """
    # Get allocation targets
    cursor = await db.execute(
        "SELECT type, name, target_pct FROM allocation_targets"
    )
    targets = {}
    for row in await cursor.fetchall():
        key = f"{row[0]}:{row[1]}"
        targets[key] = row[2]

    # Get positions with stock info
    cursor = await db.execute("""
        SELECT p.symbol, p.quantity, p.current_price, p.avg_price,
               s.name, s.geography, s.industry
        FROM positions p
        JOIN stocks s ON p.symbol = s.symbol
    """)
    positions = await cursor.fetchall()

    # Calculate totals by geography and industry
    geo_values = {}
    industry_values = {}
    total_value = 0.0

    for pos in positions:
        price = pos[2] or pos[3]  # current_price or avg_price
        value = pos[1] * price
        total_value += value

        geo = pos[5]
        industry = pos[6]

        geo_values[geo] = geo_values.get(geo, 0) + value
        industry_values[industry] = industry_values.get(industry, 0) + value

    # Get cash balance from latest snapshot
    cursor = await db.execute("""
        SELECT cash_balance FROM portfolio_snapshots
        ORDER BY date DESC LIMIT 1
    """)
    row = await cursor.fetchone()
    cash_balance = row[0] if row else 0

    # Build allocation status lists
    geo_allocations = []
    for geo in ["EU", "ASIA", "US"]:
        target = targets.get(f"geography:{geo}", 0)
        current_val = geo_values.get(geo, 0)
        current_pct = current_val / total_value if total_value > 0 else 0

        geo_allocations.append(AllocationStatus(
            category="geography",
            name=geo,
            target_pct=target,
            current_pct=round(current_pct, 4),
            current_value=round(current_val, 2),
            deviation=round(current_pct - target, 4),
        ))

    industry_allocations = []
    for industry in ["Technology", "Healthcare", "Finance", "Consumer", "Industrial"]:
        target = targets.get(f"industry:{industry}", 0)
        current_val = industry_values.get(industry, 0)
        current_pct = current_val / total_value if total_value > 0 else 0

        industry_allocations.append(AllocationStatus(
            category="industry",
            name=industry,
            target_pct=target,
            current_pct=round(current_pct, 4),
            current_value=round(current_val, 2),
            deviation=round(current_pct - target, 4),
        ))

    return PortfolioSummary(
        total_value=round(total_value, 2),
        cash_balance=round(cash_balance, 2),
        geographic_allocations=geo_allocations,
        industry_allocations=industry_allocations,
    )


async def calculate_rebalance_trades(
    db: aiosqlite.Connection,
    deposit_amount: float = None
) -> list[TradeRecommendation]:
    """
    Calculate optimal trades for rebalancing with a monthly deposit.

    Strategy:
    1. Identify underweight geographies
    2. Within underweight areas, pick highest-scored stocks
    3. Allocate deposit to bring closest to target allocations
    """
    if deposit_amount is None:
        deposit_amount = settings.monthly_deposit

    # Get current portfolio summary
    summary = await get_portfolio_summary(db)

    # Get scored stocks not currently held (or can add to)
    cursor = await db.execute("""
        SELECT s.symbol, s.name, s.geography, s.industry,
               sc.total_score, p.quantity, p.current_price
        FROM stocks s
        LEFT JOIN scores sc ON s.symbol = sc.symbol
        LEFT JOIN positions p ON s.symbol = p.symbol
        WHERE s.active = 1
        ORDER BY sc.total_score DESC NULLS LAST
    """)
    stocks = await cursor.fetchall()

    # Get current prices from Yahoo for stocks we might buy
    from app.services import yahoo

    recommendations = []
    remaining_deposit = deposit_amount

    # Sort geographies by deviation (most underweight first)
    geo_sorted = sorted(
        summary.geographic_allocations,
        key=lambda x: x.deviation
    )

    for geo_alloc in geo_sorted:
        if remaining_deposit <= 0:
            break

        # Skip if not underweight
        if geo_alloc.deviation >= 0:
            continue

        # How much do we need to invest in this geography?
        # Simple approach: proportional to underweight severity
        underweight_amount = abs(geo_alloc.deviation) * (summary.total_value + deposit_amount)
        invest_amount = min(underweight_amount, remaining_deposit * 0.5)  # Don't put all eggs in one basket

        if invest_amount < 50:  # Minimum trade size
            continue

        # Find best stocks in this geography
        geo_stocks = [
            s for s in stocks
            if s[2] == geo_alloc.name  # geography match
        ]

        if not geo_stocks:
            continue

        # Pick top scored stock in this geography
        best_stock = geo_stocks[0]  # Already sorted by score

        symbol = best_stock[0]
        name = best_stock[1]
        current_qty = best_stock[5] or 0

        # Get current price
        price = yahoo.get_current_price(symbol)
        if not price or price <= 0:
            continue

        # Calculate quantity to buy
        qty = int(invest_amount / price)
        if qty <= 0:
            continue

        actual_value = qty * price

        recommendations.append(TradeRecommendation(
            symbol=symbol,
            name=name,
            side="BUY",
            quantity=qty,
            estimated_price=round(price, 2),
            estimated_value=round(actual_value, 2),
            reason=f"Underweight {geo_alloc.name} ({geo_alloc.deviation*100:.1f}%), score: {best_stock[4]:.2f}",
        ))

        remaining_deposit -= actual_value

    # If we still have money left, diversify across remaining underweight areas
    if remaining_deposit > 100:
        # Find any stock with high score that we don't already have much of
        for stock in stocks[:5]:  # Top 5 scored stocks
            if remaining_deposit < 50:
                break

            symbol = stock[0]
            name = stock[1]
            current_qty = stock[5] or 0
            score = stock[4] or 0.5

            # Skip if already recommended
            if any(r.symbol == symbol for r in recommendations):
                continue

            price = yahoo.get_current_price(symbol)
            if not price or price <= 0:
                continue

            # Buy small amount
            invest = min(remaining_deposit * 0.3, 200)
            qty = int(invest / price)
            if qty <= 0:
                continue

            actual_value = qty * price

            recommendations.append(TradeRecommendation(
                symbol=symbol,
                name=name,
                side="BUY",
                quantity=qty,
                estimated_price=round(price, 2),
                estimated_value=round(actual_value, 2),
                reason=f"High score ({score:.2f}), diversification",
            ))

            remaining_deposit -= actual_value

    logger.info(
        f"Generated {len(recommendations)} trade recommendations, "
        f"total value: {deposit_amount - remaining_deposit:.2f}"
    )

    return recommendations


async def execute_trades(
    db: aiosqlite.Connection,
    trades: list[TradeRecommendation]
) -> list[dict]:
    """
    Execute a list of trade recommendations via Tradernet.

    Returns list of execution results.
    """
    from app.services.tradernet import get_tradernet_client
    from datetime import datetime

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
                # Record trade in database
                await db.execute(
                    """
                    INSERT INTO trades (symbol, side, quantity, price, executed_at, order_id)
                    VALUES (?, ?, ?, ?, ?, ?)
                    """,
                    (
                        trade.symbol,
                        trade.side,
                        trade.quantity,
                        result.price or trade.estimated_price,
                        datetime.now().isoformat(),
                        result.order_id,
                    ),
                )

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

    await db.commit()
    return results
