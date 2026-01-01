"""HTTP Client for Opportunity Service."""

from app.infrastructure.http_clients.base import BaseHTTPClient
from services.opportunity.models import (
    IdentifyOpportunitiesRequest,
    IdentifyOpportunitiesResponse,
)


class OpportunityHTTPClient(BaseHTTPClient):
    """HTTP client for Opportunity Service."""

    async def identify_opportunities(
        self, request: IdentifyOpportunitiesRequest
    ) -> IdentifyOpportunitiesResponse:
        """
        Identify trading opportunities.

        Args:
            request: Opportunity identification request

        Returns:
            Categorized opportunities
        """
        response = await self.post(
            "/opportunity/identify",
            json=request.model_dump(),
            timeout=10.0,
        )
        return IdentifyOpportunitiesResponse(**response.json())

    async def health_check(self) -> dict:
        """Check service health."""
        response = await self.get("/opportunity/health")
        return response.json()
