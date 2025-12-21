"""Stock universe API endpoints."""

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from typing import Optional
import aiosqlite
from app.database import get_db

router = APIRouter()


class StockCreate(BaseModel):
    """Request model for creating a stock."""
    symbol: str
    name: str
    geography: str  # EU, ASIA, US
    industry: Optional[str] = None  # Auto-detect if not provided


class StockUpdate(BaseModel):
    """Request model for updating a stock."""
    name: Optional[str] = None
    geography: Optional[str] = None
    industry: Optional[str] = None
    active: Optional[bool] = None


@router.get("")
async def get_stocks(db: aiosqlite.Connection = Depends(get_db)):
    """Get all stocks in universe with current scores and position data."""
    cursor = await db.execute("""
        SELECT s.*, sc.technical_score, sc.analyst_score,
               sc.fundamental_score, sc.total_score, sc.calculated_at,
               p.quantity as shares, p.current_price, p.avg_price,
               p.market_value_eur as position_value
        FROM stocks s
        LEFT JOIN scores sc ON s.symbol = sc.symbol
        LEFT JOIN positions p ON s.symbol = p.symbol
        WHERE s.active = 1
        ORDER BY sc.total_score DESC NULLS LAST
    """)
    rows = await cursor.fetchall()
    return [dict(row) for row in rows]


@router.get("/{symbol}")
async def get_stock(symbol: str, db: aiosqlite.Connection = Depends(get_db)):
    """Get detailed stock info with score breakdown."""
    cursor = await db.execute("""
        SELECT s.*, sc.technical_score, sc.analyst_score,
               sc.fundamental_score, sc.total_score, sc.calculated_at
        FROM stocks s
        LEFT JOIN scores sc ON s.symbol = sc.symbol
        WHERE s.symbol = ?
    """, (symbol,))
    row = await cursor.fetchone()

    if not row:
        raise HTTPException(status_code=404, detail="Stock not found")

    # Get position if any
    cursor = await db.execute("""
        SELECT * FROM positions WHERE symbol = ?
    """, (symbol,))
    position = await cursor.fetchone()

    return {
        **dict(row),
        "position": dict(position) if position else None,
    }


@router.post("/{symbol}/refresh")
async def refresh_stock_score(symbol: str, db: aiosqlite.Connection = Depends(get_db)):
    """Trigger score recalculation for a stock."""
    # Check stock exists
    cursor = await db.execute("SELECT 1 FROM stocks WHERE symbol = ?", (symbol,))
    if not await cursor.fetchone():
        raise HTTPException(status_code=404, detail="Stock not found")

    from app.services.scorer import calculate_stock_score

    score = calculate_stock_score(symbol)
    if score:
        await db.execute(
            """
            INSERT OR REPLACE INTO scores
            (symbol, technical_score, analyst_score, fundamental_score,
             total_score, calculated_at)
            VALUES (?, ?, ?, ?, ?, ?)
            """,
            (
                symbol,
                score.technical.total,
                score.analyst.total,
                score.fundamental.total,
                score.total_score,
                score.calculated_at.isoformat(),
            ),
        )
        await db.commit()

        return {
            "symbol": symbol,
            "total_score": score.total_score,
            "technical": score.technical.total,
            "analyst": score.analyst.total,
            "fundamental": score.fundamental.total,
        }

    raise HTTPException(status_code=500, detail="Failed to calculate score")


@router.post("/refresh-all")
async def refresh_all_scores(db: aiosqlite.Connection = Depends(get_db)):
    """Recalculate scores for all stocks in universe and update industries."""
    from app.services.scorer import score_all_stocks
    from app.services import yahoo

    try:
        # Get all active stocks
        cursor = await db.execute("SELECT symbol FROM stocks WHERE active = 1")
        rows = await cursor.fetchall()

        # Update industries from Yahoo Finance for stocks without industry
        for row in rows:
            symbol = row[0]
            cursor = await db.execute(
                "SELECT industry FROM stocks WHERE symbol = ?",
                (symbol,)
            )
            stock_row = await cursor.fetchone()
            if not stock_row[0]:  # No industry set
                industry = yahoo.get_stock_industry(symbol)
                if industry:
                    await db.execute(
                        "UPDATE stocks SET industry = ? WHERE symbol = ?",
                        (industry, symbol)
                    )

        await db.commit()

        # Now calculate scores
        scores = await score_all_stocks(db)
        return {
            "message": f"Refreshed scores for {len(scores)} stocks",
            "scores": [
                {"symbol": s.symbol, "total_score": s.total_score}
                for s in scores
            ],
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("")
async def create_stock(stock: StockCreate, db: aiosqlite.Connection = Depends(get_db)):
    """Add a new stock to the universe."""
    # Check if already exists
    cursor = await db.execute(
        "SELECT 1 FROM stocks WHERE symbol = ?",
        (stock.symbol.upper(),)
    )
    if await cursor.fetchone():
        raise HTTPException(status_code=400, detail="Stock already exists")

    # Validate geography
    if stock.geography.upper() not in ["EU", "ASIA", "US"]:
        raise HTTPException(
            status_code=400,
            detail="Geography must be EU, ASIA, or US"
        )

    # Auto-detect industry if not provided
    industry = stock.industry
    if not industry:
        from app.services import yahoo
        industry = yahoo.get_stock_industry(stock.symbol)

    # Insert stock
    await db.execute(
        """
        INSERT INTO stocks (symbol, name, geography, industry, active)
        VALUES (?, ?, ?, ?, 1)
        """,
        (
            stock.symbol.upper(),
            stock.name,
            stock.geography.upper(),
            industry,
        )
    )
    await db.commit()

    return {
        "message": f"Stock {stock.symbol.upper()} added to universe",
        "symbol": stock.symbol.upper(),
        "name": stock.name,
        "geography": stock.geography.upper(),
        "industry": industry,
    }


@router.put("/{symbol}")
async def update_stock(
    symbol: str,
    update: StockUpdate,
    db: aiosqlite.Connection = Depends(get_db)
):
    """Update stock details."""
    # Check stock exists
    cursor = await db.execute(
        "SELECT * FROM stocks WHERE symbol = ?",
        (symbol.upper(),)
    )
    row = await cursor.fetchone()
    if not row:
        raise HTTPException(status_code=404, detail="Stock not found")

    # Build update query
    updates = []
    values = []

    if update.name is not None:
        updates.append("name = ?")
        values.append(update.name)

    if update.geography is not None:
        if update.geography.upper() not in ["EU", "ASIA", "US"]:
            raise HTTPException(
                status_code=400,
                detail="Geography must be EU, ASIA, or US"
            )
        updates.append("geography = ?")
        values.append(update.geography.upper())

    if update.industry is not None:
        updates.append("industry = ?")
        values.append(update.industry)

    if update.active is not None:
        updates.append("active = ?")
        values.append(1 if update.active else 0)

    if not updates:
        raise HTTPException(status_code=400, detail="No updates provided")

    values.append(symbol.upper())

    await db.execute(
        f"UPDATE stocks SET {', '.join(updates)} WHERE symbol = ?",
        values
    )
    await db.commit()

    # Return updated stock
    cursor = await db.execute(
        "SELECT * FROM stocks WHERE symbol = ?",
        (symbol.upper(),)
    )
    row = await cursor.fetchone()
    return dict(row)


@router.delete("/{symbol}")
async def delete_stock(symbol: str, db: aiosqlite.Connection = Depends(get_db)):
    """Remove a stock from the universe (soft delete by setting active=0)."""
    # Check stock exists
    cursor = await db.execute(
        "SELECT 1 FROM stocks WHERE symbol = ?",
        (symbol.upper(),)
    )
    if not await cursor.fetchone():
        raise HTTPException(status_code=404, detail="Stock not found")

    # Soft delete - set active = 0
    await db.execute(
        "UPDATE stocks SET active = 0 WHERE symbol = ?",
        (symbol.upper(),)
    )
    await db.commit()

    return {"message": f"Stock {symbol.upper()} removed from universe"}
