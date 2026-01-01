"""API routes for Generator Service."""

import json
from typing import AsyncIterator

from fastapi import APIRouter, Depends
from fastapi.responses import StreamingResponse

from app.modules.planning.services.local_generator_service import (
    LocalGeneratorService,
)
from services.generator.dependencies import get_generator_service
from services.generator.models import (
    GenerateSequencesRequest,
    HealthResponse,
    SequenceBatch,
)

router = APIRouter()


async def generate_batches_stream(
    request: GenerateSequencesRequest, service: LocalGeneratorService
) -> AsyncIterator[str]:
    """
    Stream sequence batches as newline-delimited JSON.

    Yields:
        JSON lines, each containing a SequenceBatch
    """
    async for batch in service.generate_sequences_batched(request):
        yield json.dumps(batch.model_dump()) + "\n"


@router.post("/generate")
async def generate_sequences(
    request: GenerateSequencesRequest,
    service: LocalGeneratorService = Depends(get_generator_service),
):
    """
    Generate action sequences from opportunities (streaming).

    Returns batches of sequences as newline-delimited JSON stream.
    Each line is a SequenceBatch object.
    """
    return StreamingResponse(
        generate_batches_stream(request, service),
        media_type="application/x-ndjson",
    )


@router.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint."""
    return HealthResponse(
        healthy=True,
        version="1.0.0",
        status="OK",
        checks={},
    )
