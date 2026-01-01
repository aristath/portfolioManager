"""Generator Service - FastAPI application."""

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from services.generator.routes import router

app = FastAPI(
    title="Generator Service",
    description="Generates and filters action sequences from trading opportunities",
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
app.include_router(router, prefix="/generator", tags=["generator"])


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "generator",
        "version": "1.0.0",
        "status": "running",
    }


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8009,
        reload=True,
    )
