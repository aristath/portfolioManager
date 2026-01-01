"""Unit tests for LocalEvaluatorService."""

from unittest.mock import AsyncMock, patch

import pytest

from app.domain.models import Security
from app.modules.planning.domain.models import ActionCandidate
from app.modules.planning.services.local_evaluator_service import LocalEvaluatorService
from app.modules.scoring.domain.models import PortfolioContext
from services.evaluator.models import (
    ActionCandidateModel,
    EvaluateSequencesRequest,
    EvaluateSequencesResponse,
    EvaluationSettings,
    PortfolioContextInput,
    PositionInput,
    SecurityInput,
    SequenceEvaluationResult,
)


@pytest.fixture
def evaluator_service():
    """Create LocalEvaluatorService instance."""
    return LocalEvaluatorService()


@pytest.fixture
def sample_request():
    """Create sample EvaluateSequencesRequest."""
    return EvaluateSequencesRequest(
        sequences=[
            [
                ActionCandidateModel(
                    side="SELL",
                    symbol="MSFT",
                    name="Microsoft Corporation",
                    quantity=10,
                    price=400.0,
                    value_eur=4000.0,
                    currency="EUR",
                    priority=1,
                    reason="Profit taking",
                    tags=["profit_taking"],
                )
            ],
            [
                ActionCandidateModel(
                    side="BUY",
                    symbol="AAPL",
                    name="Apple Inc.",
                    quantity=10,
                    price=200.0,
                    value_eur=2000.0,
                    currency="EUR",
                    priority=1,
                    reason="Good opportunity",
                    tags=["opportunity_buy"],
                )
            ],
        ],
        portfolio_context=PortfolioContextInput(
            total_value_eur=100000.0,
            available_cash=5000.0,
            invested_value=95000.0,
            num_positions=2,
        ),
        positions=[
            PositionInput(
                symbol="AAPL",
                quantity=50,
                average_cost=180.0,
                current_price=200.0,
                value_eur=10000.0,
                currency="EUR",
                unrealized_gain_loss=1000.0,
                unrealized_gain_loss_percent=0.1,
            ),
            PositionInput(
                symbol="MSFT",
                quantity=40,
                average_cost=350.0,
                current_price=375.0,
                value_eur=15000.0,
                currency="EUR",
                unrealized_gain_loss=1000.0,
                unrealized_gain_loss_percent=0.07,
            ),
        ],
        securities=[
            SecurityInput(
                symbol="AAPL",
                name="Apple Inc.",
                current_price=200.0,
                currency="EUR",
                industry="Technology",
            ),
            SecurityInput(
                symbol="MSFT",
                name="Microsoft Corporation",
                current_price=375.0,
                currency="EUR",
                industry="Technology",
            ),
        ],
        settings=EvaluationSettings(
            beam_width=10,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
        ),
    )


@pytest.mark.asyncio
async def test_evaluate_sequences(evaluator_service, sample_request):
    """Test sequence evaluation with beam search."""
    # Mock simulate_sequence
    mock_context = PortfolioContext(
        total_value=105000.0,
        positions={"AAPL": 10000.0, "MSFT": 11000.0},
        country_weights={"US": 0.21},
        industry_weights={"Technology": 0.21},
    )

    with patch(
        "app.modules.planning.services.local_evaluator_service.simulate_sequence",
        new=AsyncMock(return_value=(mock_context, 5000.0)),
    ):
        response = await evaluator_service.evaluate_sequences(sample_request)

    assert isinstance(response, EvaluateSequencesResponse)
    assert len(response.top_sequences) == 2  # Both sequences evaluated
    assert response.total_evaluated == 2
    assert response.beam_width == 10

    # Verify sequences are sorted by total_score
    assert all(
        isinstance(seq, SequenceEvaluationResult) for seq in response.top_sequences
    )
    if len(response.top_sequences) > 1:
        assert (
            response.top_sequences[0].total_score
            >= response.top_sequences[1].total_score
        )


@pytest.mark.asyncio
async def test_evaluate_sequences_beam_limit(evaluator_service):
    """Test beam search respects beam width limit."""
    # Create request with 15 sequences but beam width of 10
    sequences = [
        [
            ActionCandidateModel(
                side="BUY",
                symbol="AAPL",
                name="Apple Inc.",
                quantity=1,
                price=200.0,
                value_eur=200.0,
                currency="EUR",
                priority=1,
                reason="Test",
                tags=[],
            )
        ]
        for _ in range(15)
    ]

    request = EvaluateSequencesRequest(
        sequences=sequences,
        portfolio_context=PortfolioContextInput(
            total_value_eur=100000.0,
            available_cash=5000.0,
            invested_value=95000.0,
            num_positions=0,
        ),
        positions=[],
        securities=[
            SecurityInput(
                symbol="AAPL",
                name="Apple Inc.",
                current_price=200.0,
                currency="EUR",
                industry="Technology",
            )
        ],
        settings=EvaluationSettings(
            beam_width=10,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
        ),
    )

    mock_context = PortfolioContext(
        total_value=100200.0,
        positions={"AAPL": 200.0},
        country_weights={"US": 0.002},
        industry_weights={"Technology": 0.002},
    )

    with patch(
        "app.modules.planning.services.local_evaluator_service.simulate_sequence",
        new=AsyncMock(return_value=(mock_context, 99800.0)),
    ):
        response = await evaluator_service.evaluate_sequences(request)

    # Should return only top 10 even though 15 were evaluated
    assert len(response.top_sequences) == 10
    assert response.total_evaluated == 15
    assert response.beam_width == 10


def test_to_portfolio_context(evaluator_service, sample_request):
    """Test conversion from request to PortfolioContext."""
    context = evaluator_service._to_portfolio_context(sample_request)

    assert isinstance(context, PortfolioContext)
    assert context.total_value == 100000.0
    assert context.positions == {"AAPL": 10000.0, "MSFT": 15000.0}


def test_to_securities(evaluator_service, sample_request):
    """Test conversion from SecurityInput to Security domain models."""
    securities = evaluator_service._to_securities(sample_request.securities)

    assert len(securities) == 2
    assert all(isinstance(s, Security) for s in securities)
    assert securities[0].symbol == "AAPL"
    assert securities[1].symbol == "MSFT"


def test_action_candidate_from_model(evaluator_service):
    """Test conversion from Pydantic model to domain ActionCandidate."""
    model = ActionCandidateModel(
        side="BUY",
        symbol="AAPL",
        name="Apple Inc.",
        quantity=10,
        price=200.0,
        value_eur=2000.0,
        currency="EUR",
        priority=1,
        reason="Test",
        tags=["test"],
    )

    action = evaluator_service._action_candidate_from_model(model)

    assert isinstance(action, ActionCandidate)
    assert action.symbol == "AAPL"
    assert action.quantity == 10
    assert action.value_eur == 2000.0


def test_action_candidate_to_model(evaluator_service):
    """Test conversion from domain ActionCandidate to Pydantic model."""
    from app.domain.value_objects.trade_side import TradeSide

    action = ActionCandidate(
        side=TradeSide.SELL,
        symbol="MSFT",
        name="Microsoft Corporation",
        quantity=5,
        price=400.0,
        value_eur=2000.0,
        currency="EUR",
        priority=1,
        reason="Test",
        tags=["test"],
    )

    model = evaluator_service._action_candidate_to_model(action)

    assert model.side == "SELL"
    assert model.symbol == "MSFT"
    assert model.quantity == 5
