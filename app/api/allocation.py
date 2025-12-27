"""Allocation target management API endpoints."""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.domain.models import AllocationTarget
from app.infrastructure.dependencies import AllocationRepositoryDep, PortfolioServiceDep

router = APIRouter()


class CountryTargets(BaseModel):
    """Dynamic country allocation weights."""

    targets: dict[str, float]


class IndustryTargets(BaseModel):
    """Dynamic industry allocation weights."""

    targets: dict[str, float]


@router.get("/targets")
async def get_allocation_targets(allocation_repo: AllocationRepositoryDep):
    """Get all allocation targets (country and industry)."""
    targets = await allocation_repo.get_all()

    country = {}
    industry = {}

    # get_all() returns Dict[str, float] with keys like "country:name" or "industry:name"
    for key, target_pct in targets.items():
        parts = key.split(":", 1)
        if len(parts) == 2:
            target_type, name = parts
            if target_type == "country":
                country[name] = target_pct
            elif target_type == "industry":
                industry[name] = target_pct

    return {
        "country": country,
        "industry": industry,
    }


@router.put("/targets/country")
async def update_country_targets(
    targets: CountryTargets, allocation_repo: AllocationRepositoryDep
):
    """Update country allocation weights."""
    updates = targets.targets

    if not updates:
        raise HTTPException(status_code=400, detail="No weights provided")

    for name, weight in updates.items():
        if weight < -1 or weight > 1:
            raise HTTPException(
                status_code=400, detail=f"Weight for {name} must be between -1 and 1"
            )

    for name, weight in updates.items():
        target = AllocationTarget(
            type="country",
            name=name,
            target_pct=weight,
        )
        await allocation_repo.upsert(target)

    country_targets = await allocation_repo.get_by_type("country")
    result = {t.name: t.target_pct for t in country_targets}

    return {
        "weights": result,
        "count": len(result),
    }


@router.put("/targets/industry")
async def update_industry_targets(
    targets: IndustryTargets, allocation_repo: AllocationRepositoryDep
):
    """Update industry allocation weights."""
    updates = targets.targets

    if not updates:
        raise HTTPException(status_code=400, detail="No weights provided")

    for name, weight in updates.items():
        if weight < -1 or weight > 1:
            raise HTTPException(
                status_code=400, detail=f"Weight for {name} must be between -1 and 1"
            )

    for name, weight in updates.items():
        target = AllocationTarget(
            type="industry",
            name=name,
            target_pct=weight,
        )
        await allocation_repo.upsert(target)

    industry_targets = await allocation_repo.get_by_type("industry")
    result = {t.name: t.target_pct for t in industry_targets}

    return {
        "weights": result,
        "count": len(result),
    }


@router.get("/current")
async def get_current_allocation(portfolio_service: PortfolioServiceDep):
    """Get current allocation vs targets for both country and industry."""
    summary = await portfolio_service.get_portfolio_summary()

    return {
        "total_value": summary.total_value,
        "cash_balance": summary.cash_balance,
        "country": [
            {
                "name": a.name,
                "target_pct": a.target_pct,
                "current_pct": a.current_pct,
                "current_value": a.current_value,
                "deviation": a.deviation,
            }
            for a in summary.geographic_allocations  # TODO: Rename to country_allocations
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
async def get_allocation_deviations(portfolio_service: PortfolioServiceDep):
    """Get allocation deviation scores for rebalancing decisions."""
    summary = await portfolio_service.get_portfolio_summary()

    country_deviations = {
        a.name: {
            "deviation": a.deviation,
            "need": max(0, -a.deviation),
            "status": (
                "underweight"
                if a.deviation < -0.02
                else ("overweight" if a.deviation > 0.02 else "balanced")
            ),
        }
        for a in summary.geographic_allocations  # TODO: Rename to country_allocations
    }

    industry_deviations = {
        a.name: {
            "deviation": a.deviation,
            "need": max(0, -a.deviation),
            "status": (
                "underweight"
                if a.deviation < -0.02
                else ("overweight" if a.deviation > 0.02 else "balanced")
            ),
        }
        for a in summary.industry_allocations
    }

    return {
        "country": country_deviations,
        "industry": industry_deviations,
    }
