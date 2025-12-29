# API routers
from app.api import (
    allocation,
    cash_flows,
    charts,
    optimizer,
    planner,
    portfolio,
    recommendations,
    status,
    stocks,
    trades,
)

__all__ = [
    "portfolio",
    "stocks",
    "trades",
    "status",
    "allocation",
    "cash_flows",
    "charts",
    "recommendations",
    "optimizer",
    "planner",
]
