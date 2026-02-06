"""Routers for Sentinel ML service."""

from sentinel_ml.api.routers.analytics import router as analytics_router
from sentinel_ml.api.routers.jobs import router as jobs_router
from sentinel_ml.api.routers.ml import router as ml_router
from sentinel_ml.api.routers.ui import router as ui_router

__all__ = ["ml_router", "analytics_router", "jobs_router", "ui_router"]
