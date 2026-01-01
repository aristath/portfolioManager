"""Local (in-process) scoring service implementation."""

from typing import List, Optional

from app.modules.scoring.services.scoring_service_interface import SecurityScore


class LocalScoringService:
    """
    Local scoring service implementation.

    Wraps existing domain logic for in-process execution.
    """

    def __init__(self):
        """Initialize local scoring service."""
        pass

    async def score_security(self, isin: str, symbol: str) -> Optional[SecurityScore]:
        """
        Score a single security.

        Args:
            isin: Security ISIN
            symbol: Security symbol

        Returns:
            Security score if found, None otherwise
        """
        # TODO: Implement using existing scoring domain logic
        return SecurityScore(
            isin=isin,
            symbol=symbol,
            total_score=0.0,
            component_scores={},
            percentile=0.0,
            grade="N/A",
        )

    async def batch_score_securities(self, isins: List[str]) -> List[SecurityScore]:
        """
        Score multiple securities.

        Args:
            isins: List of ISINs to score

        Returns:
            List of security scores
        """
        # TODO: Implement batch scoring
        return []
