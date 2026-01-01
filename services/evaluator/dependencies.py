"""Dependency injection for Evaluator Service."""

from functools import lru_cache

from app.modules.planning.services.local_evaluator_service import (
    LocalEvaluatorService,
)


@lru_cache()
def get_evaluator_service() -> LocalEvaluatorService:
    """Get singleton instance of LocalEvaluatorService."""
    return LocalEvaluatorService()
