"""Allocation target management API endpoints."""

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from typing import Optional
import aiosqlite
from app.database import get_db

router = APIRouter()


class AllocationTarget(BaseModel):
    """Single allocation target."""
    name: str
    target_pct: float  # 0.0 to 1.0


class GeographyTargets(BaseModel):
    """Geography allocation targets."""
    EU: Optional[float] = None
    ASIA: Optional[float] = None
    US: Optional[float] = None


class IndustryTargets(BaseModel):
    """
    Dynamic industry allocation targets.

    Accepts any industry names as keys with percentage values.
    Example: {"Technology": 0.20, "Defense": 0.10, "Industrial": 0.30}
    """
    targets: dict[str, float]


@router.get("/targets")
async def get_allocation_targets(db: aiosqlite.Connection = Depends(get_db)):
    """Get all allocation targets (geography and industry)."""
    cursor = await db.execute(
        "SELECT type, name, target_pct FROM allocation_targets ORDER BY type, name"
    )
    rows = await cursor.fetchall()

    geography = {}
    industry = {}

    for row in rows:
        if row["type"] == "geography":
            geography[row["name"]] = row["target_pct"]
        elif row["type"] == "industry":
            industry[row["name"]] = row["target_pct"]

    return {
        "geography": geography,
        "industry": industry,
    }


@router.put("/targets/geography")
async def update_geography_targets(
    targets: GeographyTargets,
    db: aiosqlite.Connection = Depends(get_db)
):
    """
    Update geography allocation targets.

    Targets should be decimals (e.g., 0.60 for 60%).
    Sum should equal 1.0 (100%).
    """
    updates = {}

    if targets.EU is not None:
        updates["EU"] = targets.EU
    if targets.ASIA is not None:
        updates["ASIA"] = targets.ASIA
    if targets.US is not None:
        updates["US"] = targets.US

    if not updates:
        raise HTTPException(status_code=400, detail="No targets provided")

    # Validate percentages are between 0 and 1
    for name, pct in updates.items():
        if pct < 0 or pct > 1:
            raise HTTPException(
                status_code=400,
                detail=f"Target for {name} must be between 0 and 1"
            )

    # Update targets
    for name, pct in updates.items():
        await db.execute(
            """
            INSERT OR REPLACE INTO allocation_targets (type, name, target_pct)
            VALUES ('geography', ?, ?)
            """,
            (name, pct)
        )

    await db.commit()

    # Fetch and return current targets
    cursor = await db.execute(
        "SELECT name, target_pct FROM allocation_targets WHERE type = 'geography'"
    )
    rows = await cursor.fetchall()

    result = {row["name"]: row["target_pct"] for row in rows}
    total = sum(result.values())

    return {
        "targets": result,
        "total": round(total, 4),
        "balanced": abs(total - 1.0) < 0.0001,
    }


@router.put("/targets/industry")
async def update_industry_targets(
    targets: IndustryTargets,
    db: aiosqlite.Connection = Depends(get_db)
):
    """
    Update industry allocation targets.

    Accepts dynamic industry names with percentage values.
    Targets should be decimals (e.g., 0.20 for 20%).
    Industries with 0% target will be removed from tracking.
    Sum should equal 1.0 (100%).
    """
    updates = targets.targets

    if not updates:
        raise HTTPException(status_code=400, detail="No targets provided")

    # Validate percentages are between 0 and 1
    for name, pct in updates.items():
        if pct < 0 or pct > 1:
            raise HTTPException(
                status_code=400,
                detail=f"Target for {name} must be between 0 and 1"
            )

    # Remove industries with 0% target
    to_remove = [name for name, pct in updates.items() if pct == 0]
    for name in to_remove:
        await db.execute(
            "DELETE FROM allocation_targets WHERE type = 'industry' AND name = ?",
            (name,)
        )

    # Update/insert non-zero targets
    for name, pct in updates.items():
        if pct > 0:
            await db.execute(
                """
                INSERT OR REPLACE INTO allocation_targets (type, name, target_pct)
                VALUES ('industry', ?, ?)
                """,
                (name, pct)
            )

    await db.commit()

    # Fetch and return current targets
    cursor = await db.execute(
        "SELECT name, target_pct FROM allocation_targets WHERE type = 'industry'"
    )
    rows = await cursor.fetchall()

    result = {row["name"]: row["target_pct"] for row in rows}
    total = sum(result.values())

    return {
        "targets": result,
        "total": round(total, 4),
        "balanced": abs(total - 1.0) < 0.0001,
    }


@router.get("/current")
async def get_current_allocation(db: aiosqlite.Connection = Depends(get_db)):
    """Get current allocation vs targets for both geography and industry."""
    from app.services.allocator import get_portfolio_summary

    summary = await get_portfolio_summary(db)

    return {
        "total_value": summary.total_value,
        "cash_balance": summary.cash_balance,
        "geography": [
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


@router.get("/deviations")
async def get_allocation_deviations(db: aiosqlite.Connection = Depends(get_db)):
    """
    Get allocation deviation scores for rebalancing decisions.

    Negative deviation = underweight (needs buying)
    Positive deviation = overweight
    """
    from app.services.allocator import get_portfolio_summary

    summary = await get_portfolio_summary(db)

    geo_deviations = {
        a.name: {
            "deviation": a.deviation,
            "need": max(0, -a.deviation),  # Positive value for underweight
            "status": "underweight" if a.deviation < -0.02 else (
                "overweight" if a.deviation > 0.02 else "balanced"
            ),
        }
        for a in summary.geographic_allocations
    }

    industry_deviations = {
        a.name: {
            "deviation": a.deviation,
            "need": max(0, -a.deviation),
            "status": "underweight" if a.deviation < -0.02 else (
                "overweight" if a.deviation > 0.02 else "balanced"
            ),
        }
        for a in summary.industry_allocations
    }

    return {
        "geography": geo_deviations,
        "industry": industry_deviations,
    }
