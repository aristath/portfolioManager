"""Cash flow repository - operations for cash_flows table (ledger).

DEPRECATED: This module is kept for backward compatibility during migration.
Import from app.modules.cash_flows.database.cash_flow_repository instead.
"""

# Backward compatibility re-export (temporary - will be removed in Phase 5)
from app.modules.cash_flows.database.cash_flow_repository import CashFlowRepository

__all__ = ["CashFlowRepository"]
