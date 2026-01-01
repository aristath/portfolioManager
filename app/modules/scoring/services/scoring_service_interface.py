"""Scoring service interface."""

from dataclasses import dataclass
from typing import Dict, List, Optional, Protocol


@dataclass
class SecurityScore:
    """Security score data class."""

    isin: str
    symbol: str
    total_score: float
    component_scores: Dict[str, float]
    percentile: float
    grade: str


class ScoringServiceInterface(Protocol):
    """Scoring service interface."""

    async def score_security(self, isin: str, symbol: str) -> Optional[SecurityScore]:
        """
        Score a single security.

        Args:
            isin: Security ISIN
            symbol: Security symbol

        Returns:
            Security score if found, None otherwise
        """
        ...

    async def batch_score_securities(self, isins: List[str]) -> List[SecurityScore]:
        """
        Score multiple securities.

        Args:
            isins: List of ISINs to score

        Returns:
            List of security scores
        """
        ...
