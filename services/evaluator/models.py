"""Pydantic models for Evaluator Service API."""

from typing import Dict, List, Optional

from pydantic import BaseModel, Field


class ActionCandidateModel(BaseModel):
    """Action candidate in a sequence."""

    side: str  # "BUY" or "SELL"
    symbol: str
    name: str
    quantity: int
    price: float
    value_eur: float
    currency: str
    priority: float
    reason: str
    tags: List[str] = Field(default_factory=list)


class PortfolioContextInput(BaseModel):
    """Portfolio context for evaluation."""

    total_value_eur: float
    available_cash: float
    invested_value: float
    num_positions: int
    target_allocation: Optional[Dict[str, float]] = None


class PositionInput(BaseModel):
    """Current position for evaluation."""

    symbol: str
    quantity: int
    average_cost: float
    current_price: float
    value_eur: float
    currency: str
    unrealized_gain_loss: float
    unrealized_gain_loss_percent: float


class SecurityInput(BaseModel):
    """Security information for evaluation."""

    symbol: str
    name: str
    current_price: float
    currency: str
    market_cap: Optional[float] = None
    sector: Optional[str] = None
    industry: Optional[str] = None


class EvaluationSettings(BaseModel):
    """Settings for sequence evaluation."""

    beam_width: int = Field(default=10, ge=1, le=100)
    enable_monte_carlo: bool = False
    monte_carlo_iterations: int = Field(default=100, ge=10, le=1000)
    enable_stochastic_scenarios: bool = False
    stochastic_scenarios_count: int = Field(default=5, ge=1, le=20)
    transaction_cost_fixed: float = Field(default=2.0, ge=0.0)
    transaction_cost_percent: float = Field(default=0.002, ge=0.0, le=0.1)


class EvaluateSequencesRequest(BaseModel):
    """Request to evaluate sequences."""

    sequences: List[List[ActionCandidateModel]]
    portfolio_context: PortfolioContextInput
    positions: List[PositionInput]
    securities: List[SecurityInput]
    settings: EvaluationSettings = Field(default_factory=EvaluationSettings)


class SequenceEvaluationResult(BaseModel):
    """Evaluation result for a sequence."""

    sequence: List[ActionCandidateModel]
    end_state_score: float
    diversification_score: float
    risk_score: float
    total_score: float
    total_cost: float
    cash_required: float
    feasible: bool
    metrics: Dict[str, float] = Field(default_factory=dict)


class EvaluateSequencesResponse(BaseModel):
    """Response with top evaluated sequences."""

    top_sequences: List[SequenceEvaluationResult]
    total_evaluated: int
    beam_width: int


class HealthResponse(BaseModel):
    """Health check response."""

    healthy: bool
    version: str
    status: str
    checks: Dict[str, str] = Field(default_factory=dict)
