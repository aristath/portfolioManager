"""Evaluator Service - FastAPI application."""

import os

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from services.evaluator.routes import router

app = FastAPI(
    title="Evaluator Service",
    description="Simulates and evaluates action sequences using beam search",
    version="1.0.0",
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routes
app.include_router(router, prefix="/evaluator", tags=["evaluator"])


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "evaluator",
        "version": "1.0.0",
        "status": "running",
    }


if __name__ == "__main__":
    # Port can be configured via environment variable for multiple instances
    port = int(os.getenv("EVALUATOR_PORT", "8010"))

    uvicorn.run(
        "services.evaluator.main:app",
        host="0.0.0.0",
        port=port,
        reload=True,
    )
