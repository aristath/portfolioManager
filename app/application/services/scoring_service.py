"""Scoring application service.

Orchestrates stock scoring operations.
"""

from typing import List, Optional
from datetime import datetime

from app.domain.repositories import (
    StockRepository,
    ScoreRepository,
)
from app.domain.repositories import StockScore
from app.services.scorer import (
    calculate_stock_score,
    StockScore as ScorerStockScore,
)


class ScoringService:
    """Application service for stock scoring operations."""

    def __init__(
        self,
        stock_repo: StockRepository,
        score_repo: ScoreRepository,
    ):
        self.stock_repo = stock_repo
        self.score_repo = score_repo

    async def calculate_and_save_score(
        self,
        symbol: str,
        yahoo_symbol: Optional[str] = None
    ) -> Optional[ScorerStockScore]:
        """
        Calculate stock score and save to database.

        Args:
            symbol: Stock symbol
            yahoo_symbol: Optional Yahoo Finance symbol override

        Returns:
            Calculated score or None if calculation failed
        """
        score = calculate_stock_score(symbol, yahoo_symbol)
        if score:
            # Convert to domain model
            domain_score = StockScore(
                symbol=score.symbol,
                technical_score=score.technical.total,
                analyst_score=score.analyst.total,
                fundamental_score=score.fundamental.total,
                total_score=score.total_score,
                volatility=score.technical.volatility_raw,
                calculated_at=score.calculated_at,
            )
            await self.score_repo.upsert(domain_score)

        return score

    async def score_all_stocks(self) -> List[ScorerStockScore]:
        """
        Score all active stocks in the universe and update database.

        Returns:
            List of calculated scores
        """
        stocks = await self.stock_repo.get_all_active()
        scores = []

        for stock in stocks:
            score = calculate_stock_score(stock.symbol, stock.yahoo_symbol)
            if score:
                scores.append(score)

                # Convert to domain model and save
                domain_score = StockScore(
                    symbol=score.symbol,
                    technical_score=score.technical.total,
                    analyst_score=score.analyst.total,
                    fundamental_score=score.fundamental.total,
                    total_score=score.total_score,
                    volatility=score.technical.volatility_raw,
                    calculated_at=score.calculated_at,
                )
                await self.score_repo.upsert(domain_score)

        return scores
