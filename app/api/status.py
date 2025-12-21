"""System status API endpoints."""

from datetime import datetime
from fastapi import APIRouter, Depends
import aiosqlite
from app.database import get_db
from app.config import settings

router = APIRouter()


@router.get("")
async def get_status(db: aiosqlite.Connection = Depends(get_db)):
    """Get system health and status."""
    # Get last sync time
    cursor = await db.execute("""
        SELECT date FROM portfolio_snapshots ORDER BY date DESC LIMIT 1
    """)
    row = await cursor.fetchone()
    last_sync = row["date"] if row else None

    # Get stock count
    cursor = await db.execute("SELECT COUNT(*) as count FROM stocks WHERE active = 1")
    stock_count = (await cursor.fetchone())["count"]

    # Get position count
    cursor = await db.execute("SELECT COUNT(*) as count FROM positions")
    position_count = (await cursor.fetchone())["count"]

    # Get cash balance to determine if rebalance will trigger
    cursor = await db.execute("""
        SELECT cash_balance FROM portfolio_snapshots ORDER BY date DESC LIMIT 1
    """)
    cash_row = await cursor.fetchone()
    cash_balance = cash_row["cash_balance"] if cash_row else 0

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
    from app.led.display import get_led_display

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


@router.post("/led/connect")
async def connect_led():
    """Attempt to connect to LED display."""
    from app.led.display import get_led_display

    display = get_led_display()
    success = display.connect()

    return {
        "connected": success,
        "message": "Connected to LED display" if success else "Failed to connect",
    }


@router.post("/led/test")
async def test_led():
    """Test LED display with success animation."""
    from app.led.display import get_led_display

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


@router.get("/tradernet")
async def get_tradernet_status():
    """Get Tradernet connection status."""
    from app.services.tradernet import get_tradernet_client

    client = get_tradernet_client()
    return {
        "connected": client.is_connected,
        "message": "Connected to Tradernet" if client.is_connected else "Not connected",
    }
