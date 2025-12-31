"""Logging context utilities for correlation IDs and structured logging.

DEPRECATED: This module is kept for backward compatibility during migration.
Import from app.core.logging instead.
"""

# Backward compatibility re-exports (temporary - will be removed in Phase 5)
from app.core.logging import (
    CorrelationIDFilter,
    clear_correlation_id,
    get_correlation_id,
    set_correlation_id,
    setup_correlation_logging,
)

__all__ = [
    "CorrelationIDFilter",
    "clear_correlation_id",
    "get_correlation_id",
    "set_correlation_id",
    "setup_correlation_logging",
]
