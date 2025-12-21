"""Trade execution API endpoints."""

from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
import aiosqlite
from app.database import get_db
from app.config import settings

router = APIRouter()


class TradeRequest(BaseModel):
    symbol: str
    side: str  # BUY or SELL
    quantity: float


class RebalancePreview(BaseModel):
    deposit_amount: float = None


@router.get("")
async def get_trades(
    limit: int = 50,
    db: aiosqlite.Connection = Depends(get_db)
):
    """Get trade history."""
    cursor = await db.execute("""
        SELECT t.*, s.name
        FROM trades t
        JOIN stocks s ON t.symbol = s.symbol
        ORDER BY t.executed_at DESC
        LIMIT ?
    """, (limit,))
    rows = await cursor.fetchall()
    return [dict(row) for row in rows]


@router.post("/execute")
async def execute_trade(
    trade: TradeRequest,
    db: aiosqlite.Connection = Depends(get_db)
):
    """Execute a manual trade."""
    if trade.side not in ("BUY", "SELL"):
        raise HTTPException(status_code=400, detail="Side must be BUY or SELL")

    # Check stock exists
    cursor = await db.execute("SELECT 1 FROM stocks WHERE symbol = ?", (trade.symbol,))
    if not await cursor.fetchone():
        raise HTTPException(status_code=404, detail="Stock not found")

    from app.services.tradernet import get_tradernet_client

    client = get_tradernet_client()
    if not client.is_connected:
        raise HTTPException(status_code=503, detail="Tradernet not connected")

    result = client.place_order(
        symbol=trade.symbol,
        side=trade.side,
        quantity=trade.quantity,
    )

    if result:
        # Record trade
        await db.execute(
            """
            INSERT INTO trades (symbol, side, quantity, price, executed_at, order_id)
            VALUES (?, ?, ?, ?, ?, ?)
            """,
            (
                trade.symbol,
                trade.side,
                trade.quantity,
                result.price,
                datetime.now().isoformat(),
                result.order_id,
            ),
        )
        await db.commit()

        return {
            "status": "success",
            "order_id": result.order_id,
            "symbol": trade.symbol,
            "side": trade.side,
            "quantity": trade.quantity,
            "price": result.price,
        }

    raise HTTPException(status_code=500, detail="Trade execution failed")


@router.get("/allocation")
async def get_allocation(db: aiosqlite.Connection = Depends(get_db)):
    """Get current portfolio allocation vs targets."""
    from app.services.allocator import get_portfolio_summary

    summary = await get_portfolio_summary(db)

    return {
        "total_value": summary.total_value,
        "cash_balance": summary.cash_balance,
        "geographic": [
            {
                "name": a.name,
                "target_pct": a.target_pct,
                "current_pct": a.current_pct,
                "current_value": a.current_value,
                "deviation": a.deviation,
            }
            for a in summary.geographic_allocations
        ],
        "industry": [
            {
                "name": a.name,
                "target_pct": a.target_pct,
                "current_pct": a.current_pct,
                "current_value": a.current_value,
                "deviation": a.deviation,
            }
            for a in summary.industry_allocations
        ],
    }


@router.post("/rebalance/preview")
async def preview_rebalance(
    request: RebalancePreview = None,
    db: aiosqlite.Connection = Depends(get_db)
):
    """Preview rebalance trades for deposit."""
    from app.services.allocator import calculate_rebalance_trades

    deposit = request.deposit_amount if request and request.deposit_amount else settings.monthly_deposit

    try:
        trades = await calculate_rebalance_trades(db, deposit)

        return {
            "deposit_amount": deposit,
            "total_trades": len(trades),
            "total_value": sum(t.estimated_value for t in trades),
            "trades": [
                {
                    "symbol": t.symbol,
                    "name": t.name,
                    "side": t.side,
                    "quantity": t.quantity,
                    "estimated_price": t.estimated_price,
                    "estimated_value": t.estimated_value,
                    "reason": t.reason,
                }
                for t in trades
            ],
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/rebalance/execute")
async def execute_rebalance(
    request: RebalancePreview = None,
    db: aiosqlite.Connection = Depends(get_db)
):
    """Execute monthly rebalance."""
    from app.services.allocator import calculate_rebalance_trades, execute_trades

    deposit = request.deposit_amount if request and request.deposit_amount else settings.monthly_deposit

    try:
        # Calculate trades
        trades = await calculate_rebalance_trades(db, deposit)

        if not trades:
            return {
                "status": "no_trades",
                "message": "No rebalance trades needed",
            }

        # Execute trades
        results = await execute_trades(db, trades)

        successful = sum(1 for r in results if r["status"] == "success")
        failed = sum(1 for r in results if r["status"] != "success")

        return {
            "status": "completed",
            "successful_trades": successful,
            "failed_trades": failed,
            "results": results,
        }

    except ConnectionError as e:
        raise HTTPException(status_code=503, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
