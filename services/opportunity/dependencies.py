"""Dependency injection for Opportunity Service."""

from functools import lru_cache

from app.modules.planning.services.local_opportunity_service import (
    LocalOpportunityService,
)


@lru_cache()
def get_opportunity_service() -> LocalOpportunityService:
    """Get singleton Opportunity Service instance."""
    return LocalOpportunityService()
