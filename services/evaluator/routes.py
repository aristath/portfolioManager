"""API routes for Evaluator Service."""

from fastapi import APIRouter, Depends

from app.modules.planning.services.local_evaluator_service import (
    LocalEvaluatorService,
)
from services.evaluator.dependencies import get_evaluator_service
from services.evaluator.models import (
    EvaluateSequencesRequest,
    EvaluateSequencesResponse,
    HealthResponse,
)

router = APIRouter()


@router.post("/evaluate", response_model=EvaluateSequencesResponse)
async def evaluate_sequences(
    request: EvaluateSequencesRequest,
    service: LocalEvaluatorService = Depends(get_evaluator_service),
):
    """
    Evaluate action sequences and return top K results.

    Uses beam search to maintain top sequences during evaluation.
    Simulates each sequence to calculate end state, diversification,
    and risk scores.
    """
    result = await service.evaluate_sequences(request)
    return result


@router.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint."""
    return HealthResponse(
        healthy=True,
        version="1.0.0",
        status="OK",
        checks={},
    )
