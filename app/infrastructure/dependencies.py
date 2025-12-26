"""FastAPI dependency injection container.

This module provides dependency functions for all repositories and services,
enabling proper dependency injection throughout the application.
"""

from fastapi import Depends
from typing import Annotated

from app.repositories import (
    StockRepository,
    PositionRepository,
    TradeRepository,
    ScoreRepository,
    AllocationRepository,
    CashFlowRepository,
    PortfolioRepository,
    HistoryRepository,
    SettingsRepository,
    RecommendationRepository,
    CalculationsRepository,
)
from app.domain.repositories.protocols import (
    IStockRepository,
    IPositionRepository,
    ITradeRepository,
    ISettingsRepository,
    IAllocationRepository,
)


# Repository Dependencies

def get_stock_repository() -> IStockRepository:
    """Get StockRepository instance."""
    return StockRepository()


def get_position_repository() -> IPositionRepository:
    """Get PositionRepository instance."""
    return PositionRepository()


def get_trade_repository() -> ITradeRepository:
    """Get TradeRepository instance."""
    return TradeRepository()


def get_score_repository() -> ScoreRepository:
    """Get ScoreRepository instance."""
    return ScoreRepository()


def get_allocation_repository() -> IAllocationRepository:
    """Get AllocationRepository instance."""
    return AllocationRepository()


def get_cash_flow_repository() -> CashFlowRepository:
    """Get CashFlowRepository instance."""
    return CashFlowRepository()


def get_portfolio_repository() -> PortfolioRepository:
    """Get PortfolioRepository instance."""
    return PortfolioRepository()


def get_history_repository() -> HistoryRepository:
    """Get HistoryRepository instance."""
    return HistoryRepository()


def get_settings_repository() -> ISettingsRepository:
    """Get SettingsRepository instance."""
    return SettingsRepository()


def get_recommendation_repository() -> RecommendationRepository:
    """Get RecommendationRepository instance."""
    return RecommendationRepository()


def get_calculations_repository() -> CalculationsRepository:
    """Get CalculationsRepository instance."""
    return CalculationsRepository()


# Type aliases for use in function signatures
StockRepositoryDep = Annotated[IStockRepository, Depends(get_stock_repository)]
PositionRepositoryDep = Annotated[IPositionRepository, Depends(get_position_repository)]
TradeRepositoryDep = Annotated[ITradeRepository, Depends(get_trade_repository)]
ScoreRepositoryDep = Annotated[ScoreRepository, Depends(get_score_repository)]
AllocationRepositoryDep = Annotated[IAllocationRepository, Depends(get_allocation_repository)]
CashFlowRepositoryDep = Annotated[CashFlowRepository, Depends(get_cash_flow_repository)]
PortfolioRepositoryDep = Annotated[PortfolioRepository, Depends(get_portfolio_repository)]
HistoryRepositoryDep = Annotated[HistoryRepository, Depends(get_history_repository)]
SettingsRepositoryDep = Annotated[ISettingsRepository, Depends(get_settings_repository)]
RecommendationRepositoryDep = Annotated[RecommendationRepository, Depends(get_recommendation_repository)]
CalculationsRepositoryDep = Annotated[CalculationsRepository, Depends(get_calculations_repository)]

