"""System status API endpoints."""

from datetime import datetime
from fastapi import APIRouter, Depends
from app.config import settings
from app.infrastructure.dependencies import (
    get_portfolio_repository,
    get_stock_repository,
    get_position_repository,
)
from app.domain.repositories import (
    PortfolioRepository,
    StockRepository,
    PositionRepository,
)

router = APIRouter()


@router.get("")
async def get_status(
    portfolio_repo: PortfolioRepository = Depends(get_portfolio_repository),
    stock_repo: StockRepository = Depends(get_stock_repository),
    position_repo: PositionRepository = Depends(get_position_repository),
):
    """Get system health and status."""
    # Get last sync time and cash balance from latest portfolio snapshot
    latest_snapshot = await portfolio_repo.get_latest()
    last_sync = latest_snapshot.date if latest_snapshot else None
    cash_balance = latest_snapshot.cash_balance if latest_snapshot else 0

    # Get stock count
    active_stocks = await stock_repo.get_all_active()
    stock_count = len(active_stocks)

    # Get position count
    positions = await position_repo.get_all()
    position_count = len(positions)

    return {
        "status": "healthy",
        "last_sync": last_sync,
        "stock_universe_count": stock_count,
        "active_positions": position_count,
        "cash_balance": cash_balance,
        "min_cash_threshold": settings.min_cash_threshold,
        "rebalance_ready": cash_balance >= settings.min_cash_threshold,
        "check_interval_minutes": settings.cash_check_interval_minutes,
    }


@router.get("/led")
async def get_led_status():
    """Get current LED matrix state."""
    from app.infrastructure.hardware.led_display import get_led_display

    display = get_led_display()
    state = display.get_state()

    return {
        "connected": display.is_connected,
        "mode": state.mode.value if state else "disconnected",
        "allocation": {
            "eu": state.geo_eu if state else 0,
            "asia": state.geo_asia if state else 0,
            "us": state.geo_us if state else 0,
        } if state else None,
        "system_status": state.system_status if state else "unknown",
    }


@router.get("/led/display")
async def get_led_display_state(
    position_repo: PositionRepository = Depends(get_position_repository),
):
    """
    Get display state for Arduino Bridge apps.

    Returns what the LED display should show including:
    - mode: current display mode (balance, syncing, api_call, error, etc.)
    - value: portfolio value for balance display
    - heartbeat: true if a heartbeat pulse should be shown
    - rgb_flash: RGB color array if RGB should flash
    """
    from app.infrastructure.hardware.led_display import get_led_display

    display = get_led_display()

    # Get current portfolio value from positions
    positions = await position_repo.get_all()
    portfolio_value = sum(pos.market_value_eur for pos in positions if pos.market_value_eur)

    # Update display value
    display.set_display_value(portfolio_value)

    return display.get_display_state()


@router.post("/led/connect")
async def connect_led():
    """Attempt to connect to LED display."""
    from app.infrastructure.hardware.led_display import get_led_display

    display = get_led_display()
    success = display.connect()

    return {
        "connected": success,
        "message": "Connected to LED display" if success else "Failed to connect",
    }


@router.post("/led/test")
async def test_led():
    """Test LED display with success animation."""
    from app.infrastructure.hardware.led_display import get_led_display

    display = get_led_display()
    if not display.is_connected:
        display.connect()

    if display.is_connected:
        display.show_success()
        return {"status": "success", "message": "Test animation sent"}

    return {"status": "error", "message": "LED display not connected"}


@router.post("/sync/portfolio")
async def trigger_portfolio_sync():
    """Manually trigger portfolio sync."""
    from app.jobs.daily_sync import sync_portfolio

    try:
        await sync_portfolio()
        return {"status": "success", "message": "Portfolio sync completed"}
    except Exception as e:
        return {"status": "error", "message": str(e)}


@router.post("/sync/prices")
async def trigger_price_sync():
    """Manually trigger price sync."""
    from app.jobs.daily_sync import sync_prices

    try:
        await sync_prices()
        return {"status": "success", "message": "Price sync completed"}
    except Exception as e:
        return {"status": "error", "message": str(e)}


@router.post("/sync/historical")
async def trigger_historical_sync():
    """Manually trigger historical data sync (stock prices + portfolio reconstruction)."""
    from app.jobs.historical_data_sync import sync_historical_data

    try:
        await sync_historical_data()
        return {"status": "success", "message": "Historical data sync completed"}
    except Exception as e:
        return {"status": "error", "message": str(e)}


@router.get("/tradernet")
async def get_tradernet_status():
    """Get Tradernet connection status."""
    from app.services.tradernet import get_tradernet_client

    client = get_tradernet_client()
    return {
        "connected": client.is_connected,
        "message": "Connected to Tradernet" if client.is_connected else "Not connected",
    }


@router.get("/jobs")
async def get_job_status():
    """Get status of all scheduled jobs."""
    from app.jobs.scheduler import get_job_health_status
    
    try:
        job_status = get_job_health_status()
        return {
            "status": "ok",
            "jobs": job_status,
        }
    except Exception as e:
        return {
            "status": "error",
            "message": str(e),
            "jobs": {},
        }
