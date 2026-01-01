"""REST API routes for Opportunity Service."""

from fastapi import APIRouter, Depends

from app.modules.planning.services.local_opportunity_service import (
    LocalOpportunityService,
)
from services.opportunity.dependencies import get_opportunity_service
from services.opportunity.models import (
    HealthResponse,
    IdentifyOpportunitiesRequest,
    IdentifyOpportunitiesResponse,
)

router = APIRouter()


@router.post("/identify", response_model=IdentifyOpportunitiesResponse)
async def identify_opportunities(
    request: IdentifyOpportunitiesRequest,
    service: LocalOpportunityService = Depends(get_opportunity_service),
):
    """
    Identify trading opportunities from portfolio state.

    Uses weight-based identification if target_weights provided,
    otherwise falls back to heuristic-based identification.

    Args:
        request: Portfolio context, positions, securities, and parameters
        service: Opportunity service instance

    Returns:
        Categorized opportunities (profit_taking, averaging_down, rebalance, etc.)
    """
    opportunities = await service.identify_opportunities(request)
    return opportunities


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
