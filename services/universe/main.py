"""Universe service REST API application."""

import logging
import sys
from pathlib import Path

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

# Add project root to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from services.universe.routes import router

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Create FastAPI application
app = FastAPI(
    title="Universe Service",
    description="Security universe management service",
    version="1.0.0",
)

# CORS Configuration - Multi-device deployment without authentication
# SECURITY NOTE: Wildcard origins acceptable for internal network deployment
# as documented in REST_API_SECURITY.md Phase 1 (no authentication yet).
# Credentials disabled for security (prevents CSRF attacks).
# TODO Phase 2: Add authentication and restrict origins to known devices.
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # All origins on internal network
    allow_credentials=False,  # CRITICAL: Must be False with wildcard origins
    allow_methods=["GET", "POST", "PUT", "DELETE", "PATCH"],
    allow_headers=["*"],
)

# Include router
app.include_router(router, prefix="/universe", tags=["universe"])


@app.on_event("startup")
async def startup_event():
    """Initialize service on startup."""
    logger.info("Universe service starting up...")
    logger.info("Service ready on port 8001")


@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown."""
    logger.info("Universe service shutting down...")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "services.universe.main:app",
        host="0.0.0.0",  # nosec B104
        port=8001,
        reload=True,
        log_level="info",
    )
