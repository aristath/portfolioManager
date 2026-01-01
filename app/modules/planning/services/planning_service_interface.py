"""Planning service interface."""

from dataclasses import dataclass
from typing import Any, AsyncIterator, Dict, List, Optional, Protocol

from app.modules.planning.domain.holistic_planner import HolisticPlan


@dataclass
class PlanRequest:
    """Planning request data class."""

    portfolio_hash: str
    available_cash: float
    securities: List[Any]
    positions: List[Any]
    target_weights: Optional[Dict[str, float]] = None
    parameters: Optional[Dict[str, Any]] = None


@dataclass
class PlanUpdate:
    """Planning progress update."""

    plan_id: str
    progress_pct: int
    current_step: str
    complete: bool
    plan: Optional[HolisticPlan] = None
    error: Optional[str] = None


class PlanningServiceInterface(Protocol):
    """Planning service interface."""

    async def create_plan(self, request: PlanRequest) -> AsyncIterator[PlanUpdate]:
        """
        Create a new portfolio plan.

        Args:
            request: Planning request

        Yields:
            Progress updates
        """
        ...

    async def get_plan(self, portfolio_hash: str) -> Optional[HolisticPlan]:
        """
        Get an existing plan.

        Args:
            portfolio_hash: Portfolio identifier

        Returns:
            Plan if found, None otherwise
        """
        ...
