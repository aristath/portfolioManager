"""Rate limiting middleware for API endpoints.

DEPRECATED: This module is kept for backward compatibility during migration.
Import from app.core.middleware instead.
"""

# Backward compatibility re-exports (temporary - will be removed in Phase 5)
from app.core.middleware import RateLimitMiddleware

__all__ = ["RateLimitMiddleware"]
