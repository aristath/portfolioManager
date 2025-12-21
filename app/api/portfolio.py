"""Portfolio API endpoints."""

from fastapi import APIRouter, Depends
import aiosqlite
from app.database import get_db

router = APIRouter()


@router.get("")
async def get_portfolio(db: aiosqlite.Connection = Depends(get_db)):
    """Get current portfolio positions with values."""
    cursor = await db.execute("""
        SELECT p.*, s.name as stock_name, s.industry, s.geography
        FROM positions p
        LEFT JOIN stocks s ON p.symbol = s.symbol
        ORDER BY (p.quantity * p.current_price) DESC
    """)
    rows = await cursor.fetchall()
    return [dict(row) for row in rows]


def infer_geography(symbol: str) -> str:
    """Infer geography from symbol suffix."""
    symbol = symbol.upper()
    if symbol.endswith(".GR") or symbol.endswith(".DE") or symbol.endswith(".PA"):
        return "EU"
    elif symbol.endswith(".AS") or symbol.endswith(".HK") or symbol.endswith(".T"):
        return "ASIA"
    elif symbol.endswith(".US"):
        return "US"
    return "OTHER"


@router.get("/summary")
async def get_portfolio_summary(db: aiosqlite.Connection = Depends(get_db)):
    """Get portfolio summary: total value, cash, allocation percentages."""
    # Get all positions with optional geography from stocks table
    cursor = await db.execute("""
        SELECT p.symbol, p.quantity, p.current_price, p.avg_price, s.geography
        FROM positions p
        LEFT JOIN stocks s ON p.symbol = s.symbol
    """)
    rows = await cursor.fetchall()

    # Calculate values by geography
    geo_values = {"EU": 0.0, "ASIA": 0.0, "US": 0.0}
    total_value = 0.0

    for row in rows:
        value = row["quantity"] * (row["current_price"] or row["avg_price"] or 0)
        total_value += value

        geo = row["geography"] or infer_geography(row["symbol"])
        if geo in geo_values:
            geo_values[geo] += value

    # Get latest snapshot for cash balance
    cursor = await db.execute("""
        SELECT cash_balance FROM portfolio_snapshots
        ORDER BY date DESC LIMIT 1
    """)
    row = await cursor.fetchone()
    cash_balance = row["cash_balance"] if row else 0

    return {
        "total_value": total_value,
        "cash_balance": cash_balance,
        "allocations": {
            "EU": geo_values.get("EU", 0) / total_value * 100 if total_value else 0,
            "ASIA": geo_values.get("ASIA", 0) / total_value * 100 if total_value else 0,
            "US": geo_values.get("US", 0) / total_value * 100 if total_value else 0,
        },
    }


@router.get("/history")
async def get_portfolio_history(db: aiosqlite.Connection = Depends(get_db)):
    """Get historical portfolio snapshots."""
    cursor = await db.execute("""
        SELECT * FROM portfolio_snapshots
        ORDER BY date DESC
        LIMIT 90
    """)
    rows = await cursor.fetchall()
    return [dict(row) for row in rows]
