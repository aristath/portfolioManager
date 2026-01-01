"""Unit tests for LocalGeneratorService."""

from unittest.mock import AsyncMock, patch

import pytest

from app.domain.value_objects.trade_side import TradeSide
from app.modules.planning.domain.models import ActionCandidate
from app.modules.planning.services.local_generator_service import LocalGeneratorService
from services.generator.models import (
    ActionCandidateModel,
    CombinatorialSettings,
    FeasibilitySettings,
    GenerateSequencesRequest,
    OpportunitiesInput,
    SequenceBatch,
)


@pytest.fixture
def generator_service():
    """Create LocalGeneratorService instance."""
    return LocalGeneratorService()


@pytest.fixture
def sample_request():
    """Create sample GenerateSequencesRequest."""
    return GenerateSequencesRequest(
        opportunities=OpportunitiesInput(
            profit_taking=[
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
            averaging_down=[],
            rebalance_sells=[],
            rebalance_buys=[],
            opportunity_buys=[
                ActionCandidateModel(
                    side="BUY",
                    symbol="AAPL",
                    name="Apple Inc.",
                    quantity=10,
                    price=200.0,
                    value_eur=2000.0,
                    currency="EUR",
                    priority=2,
                    reason="Good opportunity",
                    tags=["opportunity_buy"],
                )
            ],
        ),
        combinatorial=CombinatorialSettings(
            enable_weighted_combinations=True,
            max_depth=3,
        ),
        feasibility=FeasibilitySettings(
            available_cash=10000.0,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
            min_trade_value=100.0,
        ),
        batch_size=500,
    )


@pytest.mark.asyncio
async def test_generate_sequences_batched(generator_service, sample_request):
    """Test batch sequence generation."""
    # Mock the domain function to return some sequences
    mock_sequences = [
        [
            ActionCandidate(
                side=TradeSide.SELL,
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
            ActionCandidate(
                side=TradeSide.BUY,
                symbol="AAPL",
                name="Apple Inc.",
                quantity=10,
                price=200.0,
                value_eur=2000.0,
                currency="EUR",
                priority=2,
                reason="Good opportunity",
                tags=["opportunity_buy"],
            )
        ],
    ]

    with patch(
        "app.modules.planning.services.local_generator_service.generate_action_sequences",
        new=AsyncMock(return_value=mock_sequences),
    ):
        batches = []
        async for batch in generator_service.generate_sequences_batched(sample_request):
            batches.append(batch)

    assert len(batches) == 1
    assert isinstance(batches[0], SequenceBatch)
    assert batches[0].batch_number == 0
    assert len(batches[0].sequences) == 2
    assert batches[0].total_batches == 1
    assert batches[0].more_available is False


@pytest.mark.asyncio
async def test_generate_sequences_multiple_batches(generator_service, sample_request):
    """Test batch sequence generation with multiple batches."""
    # Create 1500 sequences to test batching (should create 3 batches of 500)
    mock_sequences = [
        [
            ActionCandidate(
                side=TradeSide.BUY,
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
        for _ in range(1500)
    ]

    with patch(
        "app.modules.planning.services.local_generator_service.generate_action_sequences",
        new=AsyncMock(return_value=mock_sequences),
    ):
        batches = []
        async for batch in generator_service.generate_sequences_batched(sample_request):
            batches.append(batch)

    assert len(batches) == 3
    assert batches[0].batch_number == 0
    assert batches[1].batch_number == 1
    assert batches[2].batch_number == 2
    assert len(batches[0].sequences) == 500
    assert len(batches[1].sequences) == 500
    assert len(batches[2].sequences) == 500
    assert batches[0].more_available is True
    assert batches[1].more_available is True
    assert batches[2].more_available is False


def test_apply_feasibility_filter_pass(generator_service):
    """Test feasibility filter passes valid sequences."""
    sequences = [
        [
            ActionCandidate(
                side=TradeSide.BUY,
                symbol="AAPL",
                name="Apple Inc.",
                quantity=10,
                price=200.0,
                value_eur=2000.0,
                currency="EUR",
                priority=1,
                reason="Test",
                tags=[],
            )
        ]
    ]

    feasibility = FeasibilitySettings(
        available_cash=5000.0,
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
        min_trade_value=100.0,
    )

    result = generator_service._apply_feasibility_filter(sequences, feasibility)

    assert len(result) == 1


def test_apply_feasibility_filter_fail_cash(generator_service):
    """Test feasibility filter rejects sequences requiring too much cash."""
    sequences = [
        [
            ActionCandidate(
                side=TradeSide.BUY,
                symbol="AAPL",
                name="Apple Inc.",
                quantity=100,
                price=200.0,
                value_eur=20000.0,
                currency="EUR",
                priority=1,
                reason="Test",
                tags=[],
            )
        ]
    ]

    feasibility = FeasibilitySettings(
        available_cash=5000.0,
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
        min_trade_value=100.0,
    )

    result = generator_service._apply_feasibility_filter(sequences, feasibility)

    assert len(result) == 0


def test_apply_feasibility_filter_fail_min_value(generator_service):
    """Test feasibility filter rejects sequences below minimum trade value."""
    sequences = [
        [
            ActionCandidate(
                side=TradeSide.BUY,
                symbol="AAPL",
                name="Apple Inc.",
                quantity=1,
                price=50.0,
                value_eur=50.0,
                currency="EUR",
                priority=1,
                reason="Test",
                tags=[],
            )
        ]
    ]

    feasibility = FeasibilitySettings(
        available_cash=5000.0,
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
        min_trade_value=100.0,
    )

    result = generator_service._apply_feasibility_filter(sequences, feasibility)

    assert len(result) == 0


def test_apply_feasibility_filter_with_sell(generator_service):
    """Test feasibility filter considers cash from sells."""
    sequences = [
        [
            ActionCandidate(
                side=TradeSide.SELL,
                symbol="MSFT",
                name="Microsoft Corporation",
                quantity=10,
                price=400.0,
                value_eur=4000.0,
                currency="EUR",
                priority=1,
                reason="Test",
                tags=[],
            ),
            ActionCandidate(
                side=TradeSide.BUY,
                symbol="AAPL",
                name="Apple Inc.",
                quantity=10,
                price=200.0,
                value_eur=2000.0,
                currency="EUR",
                priority=2,
                reason="Test",
                tags=[],
            ),
        ]
    ]

    feasibility = FeasibilitySettings(
        available_cash=1000.0,  # Not enough alone, but sell provides cash
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
        min_trade_value=100.0,
    )

    result = generator_service._apply_feasibility_filter(sequences, feasibility)

    # Should pass because sell generates ~3998 EUR, total available ~4998 EUR > 2002 EUR needed
    assert len(result) == 1


def test_action_candidate_from_model(generator_service):
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
        tags=["test_tag"],
    )

    action = generator_service._action_candidate_from_model(model)

    assert isinstance(action, ActionCandidate)
    assert action.side == TradeSide.BUY
    assert action.symbol == "AAPL"
    assert action.quantity == 10
    assert action.tags == ["test_tag"]


def test_action_candidate_to_model(generator_service):
    """Test conversion from domain ActionCandidate to Pydantic model."""
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
        tags=["test_tag"],
    )

    model = generator_service._action_candidate_to_model(action)

    assert isinstance(model, ActionCandidateModel)
    assert model.side == "SELL"
    assert model.symbol == "MSFT"
    assert model.quantity == 5
    assert model.tags == ["test_tag"]
