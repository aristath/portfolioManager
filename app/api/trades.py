"""Trade execution API endpoints."""

from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from app.config import settings
from app.infrastructure.dependencies import (
    get_portfolio_repository,
    get_position_repository,
    get_allocation_repository,
    get_trade_repository,
    get_stock_repository,
)
from app.domain.repositories import (
    PortfolioRepository,
    PositionRepository,
    AllocationRepository,
    TradeRepository,
    StockRepository,
)

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
    trade_repo: TradeRepository = Depends(get_trade_repository),
):
    """Get trade history."""
    trades = await trade_repo.get_history(limit=limit)
    return [
        {
            "id": None,  # Not in domain model
            "symbol": t.symbol,
            "side": t.side,
            "quantity": t.quantity,
            "price": t.price,
            "executed_at": t.executed_at.isoformat() if t.executed_at else None,
            "order_id": t.order_id,
        }
        for t in trades
    ]


@router.post("/execute")
async def execute_trade(
    trade: TradeRequest,
    stock_repo: StockRepository = Depends(get_stock_repository),
    trade_repo: TradeRepository = Depends(get_trade_repository),
):
    """Execute a manual trade."""
    if trade.side not in ("BUY", "SELL"):
        raise HTTPException(status_code=400, detail="Side must be BUY or SELL")

    # Check stock exists
    stock = await stock_repo.get_by_symbol(trade.symbol)
    if not stock:
        raise HTTPException(status_code=404, detail="Stock not found")

    from app.services.tradernet import get_tradernet_client
    from app.domain.repositories import Trade

    client = get_tradernet_client()
    if not client.is_connected:
        raise HTTPException(status_code=503, detail="Tradernet not connected")

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
            price=result.price,
            executed_at=datetime.now(),
            order_id=result.order_id,
        )
        await trade_repo.create(trade_record)

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
async def get_allocation(
    portfolio_repo: PortfolioRepository = Depends(get_portfolio_repository),
    position_repo: PositionRepository = Depends(get_position_repository),
    allocation_repo: AllocationRepository = Depends(get_allocation_repository),
):
    """Get current portfolio allocation vs targets."""
    from app.application.services.portfolio_service import PortfolioService

    portfolio_service = PortfolioService(
        portfolio_repo,
        position_repo,
        allocation_repo,
    )
    summary = await portfolio_service.get_portfolio_summary()

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
    stock_repo: StockRepository = Depends(get_stock_repository),
    position_repo: PositionRepository = Depends(get_position_repository),
    allocation_repo: AllocationRepository = Depends(get_allocation_repository),
    portfolio_repo: PortfolioRepository = Depends(get_portfolio_repository),
):
    """Preview rebalance trades for deposit."""
    from app.application.services.rebalancing_service import RebalancingService

    deposit = request.deposit_amount if request and request.deposit_amount else settings.monthly_deposit

    try:
        rebalancing_service = RebalancingService(
            stock_repo,
            position_repo,
            allocation_repo,
            portfolio_repo,
        )
        trades = await rebalancing_service.calculate_rebalance_trades(deposit)

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
    stock_repo: StockRepository = Depends(get_stock_repository),
    position_repo: PositionRepository = Depends(get_position_repository),
    allocation_repo: AllocationRepository = Depends(get_allocation_repository),
    portfolio_repo: PortfolioRepository = Depends(get_portfolio_repository),
    trade_repo: TradeRepository = Depends(get_trade_repository),
):
    """Execute monthly rebalance."""
    from app.application.services.rebalancing_service import RebalancingService
    from app.application.services.trade_execution_service import TradeExecutionService

    deposit = request.deposit_amount if request and request.deposit_amount else settings.monthly_deposit

    try:
        # Calculate trades using application service
        rebalancing_service = RebalancingService(
            stock_repo,
            position_repo,
            allocation_repo,
            portfolio_repo,
        )
        trades = await rebalancing_service.calculate_rebalance_trades(deposit)

        if not trades:
            return {
                "status": "no_trades",
                "message": "No rebalance trades needed",
            }

        # Execute trades using application service
        trade_execution_service = TradeExecutionService(trade_repo)
        results = await trade_execution_service.execute_trades(trades)

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
