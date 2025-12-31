"""Satellites API endpoints."""

from typing import List, Optional

from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field

from app.infrastructure.dependencies import (
    BalanceServiceDep,
    BucketServiceDep,
    ReconciliationServiceDep,
)
from app.modules.satellites.domain.enums import TransactionType

router = APIRouter()


# ============================================================================
# Request/Response Models
# ============================================================================


class CreateSatelliteRequest(BaseModel):
    """Request to create a new satellite bucket."""

    id: str = Field(..., description="Unique identifier for the satellite")
    name: str = Field(..., description="Human-readable name")
    notes: Optional[str] = Field(None, description="Strategy documentation")
    start_in_research: bool = Field(
        True, description="Start in research mode (paper trading)"
    )


class UpdateBucketRequest(BaseModel):
    """Request to update bucket fields."""

    name: Optional[str] = None
    notes: Optional[str] = None
    target_pct: Optional[float] = Field(None, ge=0, le=1)
    min_pct: Optional[float] = Field(None, ge=0, le=1)
    max_pct: Optional[float] = Field(None, ge=0, le=1)


class SatelliteSettingsRequest(BaseModel):
    """Request to update satellite settings."""

    preset: Optional[str] = None
    risk_appetite: float = Field(0.5, ge=0, le=1)
    hold_duration: float = Field(0.5, ge=0, le=1)
    entry_style: float = Field(0.5, ge=0, le=1)
    position_spread: float = Field(0.5, ge=0, le=1)
    profit_taking: float = Field(0.5, ge=0, le=1)
    trailing_stops: bool = False
    follow_regime: bool = False
    auto_harvest: bool = False
    pause_high_volatility: bool = False
    dividend_handling: str = "reinvest_same"


class TransferRequest(BaseModel):
    """Request to transfer cash between buckets."""

    from_bucket_id: str
    to_bucket_id: str
    amount: float = Field(..., gt=0)
    currency: str = "EUR"
    description: Optional[str] = None


class DepositRequest(BaseModel):
    """Request to allocate a deposit."""

    amount: float = Field(..., gt=0)
    currency: str = "EUR"
    description: Optional[str] = None


class ReconcileRequest(BaseModel):
    """Request to reconcile balances."""

    currency: str = "EUR"
    actual_balance: float
    auto_correct_threshold: Optional[float] = 1.0


class BucketResponse(BaseModel):
    """Response model for a bucket."""

    id: str
    name: str
    type: str
    status: str
    notes: Optional[str] = None
    target_pct: Optional[float] = None
    min_pct: Optional[float] = None
    max_pct: Optional[float] = None
    consecutive_losses: int = 0
    max_consecutive_losses: int = 5
    high_water_mark: float = 0.0
    high_water_mark_date: Optional[str] = None
    loss_streak_paused_at: Optional[str] = None
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    class Config:
        from_attributes = True


class BalanceResponse(BaseModel):
    """Response model for a bucket balance."""

    bucket_id: str
    currency: str
    balance: float
    last_updated: str


class TransactionResponse(BaseModel):
    """Response model for a transaction."""

    id: Optional[int] = None
    bucket_id: str
    type: str
    amount: float
    currency: str
    description: Optional[str] = None
    created_at: Optional[str] = None


class SettingsResponse(BaseModel):
    """Response model for satellite settings."""

    satellite_id: str
    preset: Optional[str] = None
    risk_appetite: float = 0.5
    hold_duration: float = 0.5
    entry_style: float = 0.5
    position_spread: float = 0.5
    profit_taking: float = 0.5
    trailing_stops: bool = False
    follow_regime: bool = False
    auto_harvest: bool = False
    pause_high_volatility: bool = False
    dividend_handling: str = "reinvest_same"

    class Config:
        from_attributes = True


class ReconciliationResultResponse(BaseModel):
    """Response model for reconciliation result."""

    currency: str
    virtual_total: float
    actual_total: float
    difference: float
    is_reconciled: bool
    adjustments_made: dict
    timestamp: str


# ============================================================================
# Bucket CRUD Endpoints
# ============================================================================


@router.get("/buckets", response_model=List[BucketResponse])
async def list_buckets(bucket_service: BucketServiceDep):
    """List all buckets (core and satellites)."""
    buckets = await bucket_service.get_all_buckets()
    return [
        BucketResponse(
            id=b.id,
            name=b.name,
            type=b.type.value,
            status=b.status.value,
            notes=b.notes,
            target_pct=b.target_pct,
            min_pct=b.min_pct,
            max_pct=b.max_pct,
            consecutive_losses=b.consecutive_losses,
            max_consecutive_losses=b.max_consecutive_losses,
            high_water_mark=b.high_water_mark,
            high_water_mark_date=b.high_water_mark_date,
            loss_streak_paused_at=b.loss_streak_paused_at,
            created_at=b.created_at,
            updated_at=b.updated_at,
        )
        for b in buckets
    ]


@router.get("/buckets/{bucket_id}", response_model=BucketResponse)
async def get_bucket(bucket_id: str, bucket_service: BucketServiceDep):
    """Get a specific bucket by ID."""
    bucket = await bucket_service.get_bucket(bucket_id)
    if not bucket:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Bucket '{bucket_id}' not found",
        )
    return BucketResponse(
        id=bucket.id,
        name=bucket.name,
        type=bucket.type.value,
        status=bucket.status.value,
        notes=bucket.notes,
        target_pct=bucket.target_pct,
        min_pct=bucket.min_pct,
        max_pct=bucket.max_pct,
        consecutive_losses=bucket.consecutive_losses,
        max_consecutive_losses=bucket.max_consecutive_losses,
        high_water_mark=bucket.high_water_mark,
        high_water_mark_date=bucket.high_water_mark_date,
        loss_streak_paused_at=bucket.loss_streak_paused_at,
        created_at=bucket.created_at,
        updated_at=bucket.updated_at,
    )


@router.post(
    "/satellites", response_model=BucketResponse, status_code=status.HTTP_201_CREATED
)
async def create_satellite(
    request: CreateSatelliteRequest, bucket_service: BucketServiceDep
):
    """Create a new satellite bucket."""
    try:
        bucket = await bucket_service.create_satellite(
            satellite_id=request.id,
            name=request.name,
            notes=request.notes,
            start_in_research=request.start_in_research,
        )
        return BucketResponse(
            id=bucket.id,
            name=bucket.name,
            type=bucket.type.value,
            status=bucket.status.value,
            notes=bucket.notes,
            target_pct=bucket.target_pct,
            min_pct=bucket.min_pct,
            max_pct=bucket.max_pct,
            consecutive_losses=bucket.consecutive_losses,
            max_consecutive_losses=bucket.max_consecutive_losses,
            high_water_mark=bucket.high_water_mark,
            high_water_mark_date=bucket.high_water_mark_date,
            loss_streak_paused_at=bucket.loss_streak_paused_at,
            created_at=bucket.created_at,
            updated_at=bucket.updated_at,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.patch("/buckets/{bucket_id}", response_model=BucketResponse)
async def update_bucket(
    bucket_id: str, request: UpdateBucketRequest, bucket_service: BucketServiceDep
):
    """Update a bucket's fields."""
    updates = request.model_dump(exclude_unset=True)
    if not updates:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="No fields to update",
        )

    bucket = await bucket_service.update_bucket(bucket_id, **updates)
    if not bucket:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Bucket '{bucket_id}' not found",
        )
    return BucketResponse(
        id=bucket.id,
        name=bucket.name,
        type=bucket.type.value,
        status=bucket.status.value,
        notes=bucket.notes,
        target_pct=bucket.target_pct,
        min_pct=bucket.min_pct,
        max_pct=bucket.max_pct,
        consecutive_losses=bucket.consecutive_losses,
        max_consecutive_losses=bucket.max_consecutive_losses,
        high_water_mark=bucket.high_water_mark,
        high_water_mark_date=bucket.high_water_mark_date,
        loss_streak_paused_at=bucket.loss_streak_paused_at,
        created_at=bucket.created_at,
        updated_at=bucket.updated_at,
    )


# ============================================================================
# Lifecycle Endpoints
# ============================================================================


@router.post("/satellites/{satellite_id}/activate", response_model=BucketResponse)
async def activate_satellite(satellite_id: str, bucket_service: BucketServiceDep):
    """Activate a satellite from research or accumulating mode."""
    try:
        bucket = await bucket_service.activate_satellite(satellite_id)
        return BucketResponse(
            id=bucket.id,
            name=bucket.name,
            type=bucket.type.value,
            status=bucket.status.value,
            notes=bucket.notes,
            target_pct=bucket.target_pct,
            min_pct=bucket.min_pct,
            max_pct=bucket.max_pct,
            consecutive_losses=bucket.consecutive_losses,
            max_consecutive_losses=bucket.max_consecutive_losses,
            high_water_mark=bucket.high_water_mark,
            high_water_mark_date=bucket.high_water_mark_date,
            loss_streak_paused_at=bucket.loss_streak_paused_at,
            created_at=bucket.created_at,
            updated_at=bucket.updated_at,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/buckets/{bucket_id}/pause", response_model=BucketResponse)
async def pause_bucket(bucket_id: str, bucket_service: BucketServiceDep):
    """Pause a bucket, stopping all trading."""
    try:
        bucket = await bucket_service.pause_bucket(bucket_id)
        return BucketResponse(
            id=bucket.id,
            name=bucket.name,
            type=bucket.type.value,
            status=bucket.status.value,
            notes=bucket.notes,
            target_pct=bucket.target_pct,
            min_pct=bucket.min_pct,
            max_pct=bucket.max_pct,
            consecutive_losses=bucket.consecutive_losses,
            max_consecutive_losses=bucket.max_consecutive_losses,
            high_water_mark=bucket.high_water_mark,
            high_water_mark_date=bucket.high_water_mark_date,
            loss_streak_paused_at=bucket.loss_streak_paused_at,
            created_at=bucket.created_at,
            updated_at=bucket.updated_at,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/buckets/{bucket_id}/resume", response_model=BucketResponse)
async def resume_bucket(bucket_id: str, bucket_service: BucketServiceDep):
    """Resume a paused bucket."""
    try:
        bucket = await bucket_service.resume_bucket(bucket_id)
        return BucketResponse(
            id=bucket.id,
            name=bucket.name,
            type=bucket.type.value,
            status=bucket.status.value,
            notes=bucket.notes,
            target_pct=bucket.target_pct,
            min_pct=bucket.min_pct,
            max_pct=bucket.max_pct,
            consecutive_losses=bucket.consecutive_losses,
            max_consecutive_losses=bucket.max_consecutive_losses,
            high_water_mark=bucket.high_water_mark,
            high_water_mark_date=bucket.high_water_mark_date,
            loss_streak_paused_at=bucket.loss_streak_paused_at,
            created_at=bucket.created_at,
            updated_at=bucket.updated_at,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/satellites/{satellite_id}/retire", response_model=BucketResponse)
async def retire_satellite(satellite_id: str, bucket_service: BucketServiceDep):
    """Retire a satellite permanently."""
    try:
        bucket = await bucket_service.retire_satellite(satellite_id)
        return BucketResponse(
            id=bucket.id,
            name=bucket.name,
            type=bucket.type.value,
            status=bucket.status.value,
            notes=bucket.notes,
            target_pct=bucket.target_pct,
            min_pct=bucket.min_pct,
            max_pct=bucket.max_pct,
            consecutive_losses=bucket.consecutive_losses,
            max_consecutive_losses=bucket.max_consecutive_losses,
            high_water_mark=bucket.high_water_mark,
            high_water_mark_date=bucket.high_water_mark_date,
            loss_streak_paused_at=bucket.loss_streak_paused_at,
            created_at=bucket.created_at,
            updated_at=bucket.updated_at,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


# ============================================================================
# Settings Endpoints
# ============================================================================


@router.get("/satellites/{satellite_id}/settings", response_model=SettingsResponse)
async def get_satellite_settings(satellite_id: str, bucket_service: BucketServiceDep):
    """Get settings for a satellite."""
    settings = await bucket_service.get_settings(satellite_id)
    if not settings:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Settings for satellite '{satellite_id}' not found",
        )
    return SettingsResponse(
        satellite_id=settings.satellite_id,
        preset=settings.preset,
        risk_appetite=settings.risk_appetite,
        hold_duration=settings.hold_duration,
        entry_style=settings.entry_style,
        position_spread=settings.position_spread,
        profit_taking=settings.profit_taking,
        trailing_stops=settings.trailing_stops,
        follow_regime=settings.follow_regime,
        auto_harvest=settings.auto_harvest,
        pause_high_volatility=settings.pause_high_volatility,
        dividend_handling=settings.dividend_handling,
    )


@router.put("/satellites/{satellite_id}/settings", response_model=SettingsResponse)
async def update_satellite_settings(
    satellite_id: str,
    request: SatelliteSettingsRequest,
    bucket_service: BucketServiceDep,
):
    """Update settings for a satellite."""
    from app.modules.satellites.domain.models import SatelliteSettings

    try:
        settings = SatelliteSettings(
            satellite_id=satellite_id,
            preset=request.preset,
            risk_appetite=request.risk_appetite,
            hold_duration=request.hold_duration,
            entry_style=request.entry_style,
            position_spread=request.position_spread,
            profit_taking=request.profit_taking,
            trailing_stops=request.trailing_stops,
            follow_regime=request.follow_regime,
            auto_harvest=request.auto_harvest,
            pause_high_volatility=request.pause_high_volatility,
            dividend_handling=request.dividend_handling,
        )
        saved = await bucket_service.save_settings(settings)
        return SettingsResponse(
            satellite_id=saved.satellite_id,
            preset=saved.preset,
            risk_appetite=saved.risk_appetite,
            hold_duration=saved.hold_duration,
            entry_style=saved.entry_style,
            position_spread=saved.position_spread,
            profit_taking=saved.profit_taking,
            trailing_stops=saved.trailing_stops,
            follow_regime=saved.follow_regime,
            auto_harvest=saved.auto_harvest,
            pause_high_volatility=saved.pause_high_volatility,
            dividend_handling=saved.dividend_handling,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


# ============================================================================
# Balance Endpoints
# ============================================================================


@router.get("/buckets/{bucket_id}/balances", response_model=List[BalanceResponse])
async def get_bucket_balances(bucket_id: str, balance_service: BalanceServiceDep):
    """Get all currency balances for a bucket."""
    balances = await balance_service.get_all_balances(bucket_id)
    return [
        BalanceResponse(
            bucket_id=b.bucket_id,
            currency=b.currency,
            balance=b.balance,
            last_updated=b.last_updated,
        )
        for b in balances
    ]


@router.get("/balances/summary")
async def get_balance_summary(balance_service: BalanceServiceDep):
    """Get summary of all bucket balances."""
    return await balance_service.get_portfolio_summary()


@router.post("/balances/transfer")
async def transfer_between_buckets(
    request: TransferRequest, balance_service: BalanceServiceDep
):
    """Transfer cash between buckets."""
    try:
        from_balance, to_balance = await balance_service.transfer_between_buckets(
            from_bucket_id=request.from_bucket_id,
            to_bucket_id=request.to_bucket_id,
            amount=request.amount,
            currency=request.currency,
            description=request.description,
        )
        return {
            "from_balance": BalanceResponse(
                bucket_id=from_balance.bucket_id,
                currency=from_balance.currency,
                balance=from_balance.balance,
                last_updated=from_balance.last_updated,
            ),
            "to_balance": BalanceResponse(
                bucket_id=to_balance.bucket_id,
                currency=to_balance.currency,
                balance=to_balance.balance,
                last_updated=to_balance.last_updated,
            ),
        }
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/balances/deposit")
async def allocate_deposit(request: DepositRequest, balance_service: BalanceServiceDep):
    """Allocate a new deposit across buckets."""
    allocations = await balance_service.allocate_deposit(
        total_amount=request.amount,
        currency=request.currency,
        description=request.description,
    )
    return {"allocations": allocations}


# ============================================================================
# Transaction History Endpoints
# ============================================================================


@router.get(
    "/buckets/{bucket_id}/transactions", response_model=List[TransactionResponse]
)
async def get_bucket_transactions(
    bucket_id: str,
    balance_service: BalanceServiceDep,
    limit: int = 100,
    transaction_type: Optional[str] = None,
):
    """Get transaction history for a bucket."""
    tx_type = TransactionType(transaction_type) if transaction_type else None
    transactions = await balance_service.get_transactions(
        bucket_id=bucket_id, limit=limit, transaction_type=tx_type
    )
    return [
        TransactionResponse(
            id=t.id,
            bucket_id=t.bucket_id,
            type=t.type.value,
            amount=t.amount,
            currency=t.currency,
            description=t.description,
            created_at=t.created_at,
        )
        for t in transactions
    ]


# ============================================================================
# Reconciliation Endpoints
# ============================================================================


@router.post("/reconcile", response_model=ReconciliationResultResponse)
async def reconcile_balances(
    request: ReconcileRequest, reconciliation_service: ReconciliationServiceDep
):
    """Reconcile virtual balances with actual brokerage balance."""
    result = await reconciliation_service.reconcile(
        currency=request.currency,
        actual_balance=request.actual_balance,
        auto_correct_threshold=request.auto_correct_threshold,
    )
    return ReconciliationResultResponse(
        currency=result.currency,
        virtual_total=result.virtual_total,
        actual_total=result.actual_total,
        difference=result.difference,
        is_reconciled=result.is_reconciled,
        adjustments_made=result.adjustments_made,
        timestamp=result.timestamp,
    )


@router.get("/reconcile/{currency}/check", response_model=ReconciliationResultResponse)
async def check_reconciliation(
    currency: str,
    actual_balance: float,
    reconciliation_service: ReconciliationServiceDep,
):
    """Check reconciliation without making changes."""
    result = await reconciliation_service.check_invariant(
        currency=currency, actual_balance=actual_balance
    )
    return ReconciliationResultResponse(
        currency=result.currency,
        virtual_total=result.virtual_total,
        actual_total=result.actual_total,
        difference=result.difference,
        is_reconciled=result.is_reconciled,
        adjustments_made=result.adjustments_made,
        timestamp=result.timestamp,
    )


@router.get("/reconcile/{currency}/breakdown")
async def get_balance_breakdown(
    currency: str, reconciliation_service: ReconciliationServiceDep
):
    """Get detailed breakdown of virtual balances by bucket."""
    return await reconciliation_service.get_balance_breakdown(currency)


# ============================================================================
# Allocation Settings Endpoints
# ============================================================================


@router.get("/settings/allocation")
async def get_allocation_settings(balance_service: BalanceServiceDep):
    """Get global allocation settings."""
    return await balance_service.get_allocation_settings()


@router.put("/settings/satellite-budget")
async def update_satellite_budget(
    budget_pct: float, balance_service: BalanceServiceDep
):
    """Update the global satellite budget percentage."""
    try:
        await balance_service.update_satellite_budget(budget_pct)
        return {"satellite_budget_pct": budget_pct}
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))
