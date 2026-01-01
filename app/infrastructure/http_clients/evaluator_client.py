"""HTTP Client for Evaluator Service."""

from app.infrastructure.http_clients.base import BaseHTTPClient
from services.evaluator.models import (
    EvaluateSequencesRequest,
    EvaluateSequencesResponse,
)


class EvaluatorHTTPClient(BaseHTTPClient):
    """HTTP client for Evaluator Service."""

    async def evaluate_sequences(
        self, request: EvaluateSequencesRequest
    ) -> EvaluateSequencesResponse:
        """
        Evaluate action sequences.

        Args:
            request: Sequence evaluation request

        Returns:
            Top K evaluated sequences with scores
        """
        response = await self.post(
            "/evaluator/evaluate",
            json=request.model_dump(),
            timeout=120.0,  # Evaluation is slow
        )
        return EvaluateSequencesResponse(**response.json())

    async def health_check(self) -> dict:
        """Check service health."""
        response = await self.get("/evaluator/health")
        return response.json()
