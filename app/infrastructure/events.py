"""Event system for decoupled LED and system notifications.

DEPRECATED: This module is kept for backward compatibility during migration.
Import from app.core.events instead.
"""

# Backward compatibility re-exports (temporary - will be removed in Phase 5)
from app.core.events import SystemEvent, emit, subscribe

__all__ = ["SystemEvent", "emit", "subscribe"]
