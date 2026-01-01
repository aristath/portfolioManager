"""HTTP Client for Generator Service."""

import json
from typing import AsyncIterator

import httpx

from app.infrastructure.http_clients.base import BaseHTTPClient
from services.generator.models import GenerateSequencesRequest, SequenceBatch


class GeneratorHTTPClient(BaseHTTPClient):
    """HTTP client for Generator Service."""

    async def generate_sequences_streaming(
        self, request: GenerateSequencesRequest
    ) -> AsyncIterator[SequenceBatch]:
        """
        Generate sequences from opportunities (streaming).

        Args:
            request: Sequence generation request

        Yields:
            SequenceBatch objects from the stream
        """
        async with httpx.AsyncClient(timeout=30.0) as client:
            async with client.stream(
                "POST",
                f"{self.base_url}/generator/generate",
                json=request.model_dump(),
            ) as response:
                response.raise_for_status()
                async for line in response.aiter_lines():
                    if line.strip():
                        batch_data = json.loads(line)
                        yield SequenceBatch(**batch_data)

    async def health_check(self) -> dict:
        """Check service health."""
        response = await self.get("/generator/health")
        return response.json()
