"""REST API routes for Planning service."""

from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, Query

from app.modules.planning.services.local_planning_service import LocalPlanningService
from app.modules.planning.services.planning_service_interface import PlanRequest
from app.modules.portfolio.domain.models import Position
from services.planning.dependencies import get_planning_service
from services.planning.models import (
    CreatePlanRequest,
    CreatePlanResponse,
    GetBestResultResponse,
    GetPlanResponse,
    HealthResponse,
    ListPlansResponse,
    Plan,
    PlannedAction,
)

router = APIRouter()


def _convert_plan_to_response(plan) -> Plan:
    """Convert domain plan to Plan response model."""
    actions = []
    for step in plan.steps:
        action = PlannedAction(
            side=step.side,
            symbol=step.symbol,
            isin="",  # HolisticStep doesn't have ISIN
            quantity=step.quantity,
            estimated_price=step.estimated_price,
            estimated_cost=abs(step.estimated_value),
            reason=step.reason,
            priority=step.step_number,
        )
        actions.append(action)

    total_value = sum(abs(step.estimated_value) for step in plan.steps)

    return Plan(
        id=str(hash(plan.narrative_summary))[:16],
        portfolio_hash="",
        actions=actions,
        score=plan.end_state_score,
        expected_cost=plan.cash_required,
        expected_value=total_value,
        created_at=datetime.now().isoformat(),
        status="READY" if plan.feasible else "DRAFT",
    )


@router.post("/create", response_model=CreatePlanResponse)
async def create_plan(
    request: CreatePlanRequest,
    service: LocalPlanningService = Depends(get_planning_service),
):
    """
    Create a new portfolio plan.

    Args:
        request: Plan creation request
        service: Planning service instance

    Returns:
        Created plan

    Note:
        This is a synchronous endpoint - streaming has been converted to blocking operation
    """
    # Convert request positions to domain positions
    positions = []
    for pos_input in request.positions:
        position = Position(
            symbol=pos_input.symbol,
            quantity=pos_input.quantity,
            avg_price=pos_input.average_price,
            isin=pos_input.isin,
            currency=None,
            currency_rate=1.0,
            current_price=pos_input.current_price,
            market_value_eur=pos_input.market_value,
            cost_basis_eur=pos_input.average_price * pos_input.quantity,
            unrealized_pnl=(pos_input.current_price - pos_input.average_price)
            * pos_input.quantity,
            unrealized_pnl_pct=(
                (pos_input.current_price - pos_input.average_price)
                / pos_input.average_price
                * 100
                if pos_input.average_price > 0
                else 0.0
            ),
            last_updated=datetime.now().isoformat(),
        )
        positions.append(position)

    # Extract target weights from constraints
    target_weights = None
    if request.constraints:
        target_weights = {}
        for key, value in request.constraints.items():
            if key.startswith("target_"):
                symbol = key.replace("target_", "")
                target_weights[symbol] = float(value)

    # Build domain request
    domain_request = PlanRequest(
        portfolio_hash=request.portfolio_hash,
        available_cash=request.available_cash,
        securities=[],  # Will be fetched by service
        positions=positions,
        target_weights=target_weights,
        parameters=request.constraints,
    )

    # Call service and collect all updates (streaming converted to blocking)
    plan_id = ""
    final_plan = None
    error = None

    try:
        async for update in service.create_plan(domain_request):
            plan_id = update.plan_id
            if update.complete:
                if update.plan:
                    final_plan = update.plan
                if update.error:
                    error = update.error
                break
    except Exception as e:
        error = str(e)

    if error:
        return CreatePlanResponse(
            plan_id=plan_id,
            success=False,
            message=error,
            plan=None,
        )

    if final_plan:
        plan_response = _convert_plan_to_response(final_plan)
        return CreatePlanResponse(
            plan_id=plan_id,
            success=True,
            message="Plan created successfully",
            plan=plan_response,
        )
    else:
        return CreatePlanResponse(
            plan_id=plan_id,
            success=False,
            message="Plan creation failed - no plan returned",
            plan=None,
        )


@router.get("/plans/{plan_id}", response_model=GetPlanResponse)
async def get_plan(
    plan_id: str,
    service: LocalPlanningService = Depends(get_planning_service),
):
    """
    Get an existing plan by ID.

    Args:
        plan_id: Plan identifier
        service: Planning service instance

    Returns:
        Plan details if found
    """
    plan = await service.get_plan(plan_id)

    if plan:
        plan_response = _convert_plan_to_response(plan)
        return GetPlanResponse(found=True, plan=plan_response)
    else:
        return GetPlanResponse(found=False)


@router.get("/plans", response_model=ListPlansResponse)
async def list_plans(
    portfolio_hash: str = Query(..., description="Portfolio hash"),
    limit: int = Query(default=100, ge=1, le=1000, description="Maximum results"),
    offset: int = Query(default=0, ge=0, description="Results offset"),
    service: LocalPlanningService = Depends(get_planning_service),
):
    """
    List all plans for a portfolio.

    Args:
        portfolio_hash: Portfolio hash
        limit: Maximum number of results
        offset: Results offset for pagination
        service: Planning service instance

    Returns:
        List of plans
    """
    # Get plans from repository
    plans = await service.planner_repo.get_plans_for_portfolio(
        portfolio_hash,
        limit=limit,
        offset=offset,
    )

    # Convert to response
    plan_responses = [_convert_plan_to_response(plan) for plan in plans]

    return ListPlansResponse(plans=plan_responses, total=len(plan_responses))


@router.get("/best", response_model=GetBestResultResponse)
async def get_best_result(
    portfolio_hash: str = Query(..., description="Portfolio hash"),
    service: LocalPlanningService = Depends(get_planning_service),
):
    """
    Get best plan result for a portfolio.

    Args:
        portfolio_hash: Portfolio hash
        service: Planning service instance

    Returns:
        Best plan if found
    """
    # Get best plan for portfolio from repository
    plans = await service.planner_repo.get_plans_for_portfolio(
        portfolio_hash,
        limit=1,
        offset=0,
    )

    if plans:
        best_plan = plans[0]  # Already sorted by score
        plan_response = _convert_plan_to_response(best_plan)
        return GetBestResultResponse(found=True, plan=plan_response)
    else:
        return GetBestResultResponse(found=False)


@router.get("/health", response_model=HealthResponse)
async def health_check():
    """
    Health check endpoint.

    Returns:
        Service health status
    """
    return HealthResponse(
        healthy=True,
        version="1.0.0",
        status="OK",
        checks={},
    )
