"""API routers for Sentinel.

Each router handles a specific domain of the API.
"""

from sentinel.api.routers.portfolio import allocation_router, targets_router
from sentinel.api.routers.portfolio import router as portfolio_router
from sentinel.api.routers.settings import led_router
from sentinel.api.routers.settings import router as settings_router

__all__ = [
    "settings_router",
    "led_router",
    "portfolio_router",
    "allocation_router",
    "targets_router",
]
