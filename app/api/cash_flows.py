"""Cash flows API endpoints.

DEPRECATED: This module is kept for backward compatibility during migration.
Import from app.modules.cash_flows.api.cash_flows instead.
"""

# Backward compatibility re-export (temporary - will be removed in Phase 5)
from app.modules.cash_flows.api.cash_flows import (
    get_cash_flows,
    get_cash_flows_summary,
    router,
    sync_cash_flows,
)

__all__ = ["router", "get_cash_flows", "sync_cash_flows", "get_cash_flows_summary"]
