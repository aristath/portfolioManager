"""Local Opportunity Service - Domain service wrapper for opportunity identification."""

from app.modules.planning.domain.models import ActionCandidate
from services.opportunity.models import (
    ActionCandidateModel,
    IdentifyOpportunitiesRequest,
    IdentifyOpportunitiesResponse,
)


class LocalOpportunityService:
    """
    Service for identifying trading opportunities.

    Wraps the opportunity identification logic from holistic_planner.py
    for use by the Opportunity microservice.
    """

    def __init__(self):
        """Initialize the service."""
        pass

    async def identify_opportunities(
        self, request: IdentifyOpportunitiesRequest
    ) -> IdentifyOpportunitiesResponse:
        """
        Identify trading opportunities from portfolio state.

        Uses weight-based identification if target_weights provided,
        otherwise falls back to heuristic-based identification.

        Args:
            request: Portfolio context, positions, securities, and parameters

        Returns:
            Categorized opportunities

        TODO: Extract logic from holistic_planner.py:
            - identify_opportunities_from_weights() (lines 382-489)
            - identify_opportunities() (lines 492-600)
            - Helper functions (lines 92-379)
        """
        # TODO: Implement opportunity identification logic
        # For now, return empty opportunities
        return IdentifyOpportunitiesResponse(
            profit_taking=[],
            averaging_down=[],
            rebalance_sells=[],
            rebalance_buys=[],
            opportunity_buys=[],
        )

    def _action_candidate_to_model(
        self, action: ActionCandidate
    ) -> ActionCandidateModel:
        """
        Convert domain ActionCandidate to Pydantic model.

        Args:
            action: Domain ActionCandidate

        Returns:
            ActionCandidateModel for API response
        """
        return ActionCandidateModel(
            side=action.side.value if hasattr(action.side, "value") else action.side,
            symbol=action.symbol,
            name=action.name,
            quantity=action.quantity,
            price=action.price,
            value_eur=action.value_eur,
            currency=action.currency,
            priority=action.priority,
            reason=action.reason,
            tags=action.tags if hasattr(action, "tags") else [],
        )
