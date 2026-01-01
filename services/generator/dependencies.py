"""Dependency injection for Generator Service."""

from functools import lru_cache

from app.modules.planning.services.local_generator_service import (
    LocalGeneratorService,
)


@lru_cache()
def get_generator_service() -> LocalGeneratorService:
    """Get singleton instance of LocalGeneratorService."""
    return LocalGeneratorService()
