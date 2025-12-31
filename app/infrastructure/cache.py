"""Simple in-memory cache with TTL.

DEPRECATED: This module is kept for backward compatibility during migration.
Import from app.core.cache instead.
"""

# Backward compatibility re-exports (temporary - will be removed in Phase 5)
from app.core.cache import SimpleCache, cache

__all__ = ["SimpleCache", "cache"]
