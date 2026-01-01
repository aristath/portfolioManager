"""Opportunity Service - Main application entrypoint."""

import logging
import sys
from pathlib import Path

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

# Add project root to path
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from services.opportunity.routes import router

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Create FastAPI application
app = FastAPI(
    title="Opportunity Service",
    description="Identifies trading opportunities from portfolio state",
    version="1.0.0",
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include router
app.include_router(router, prefix="/opportunity", tags=["opportunity"])


@app.on_event("startup")
async def startup_event():
    """Log startup message."""
    logger.info("Opportunity service starting up...")
    logger.info("Service ready on port 8008")


@app.on_event("shutdown")
async def shutdown_event():
    """Log shutdown message."""
    logger.info("Opportunity service shutting down...")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=8008, reload=True)
