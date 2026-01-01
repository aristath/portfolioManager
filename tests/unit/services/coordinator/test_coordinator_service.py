"""Unit tests for LocalCoordinatorService."""

from unittest.mock import AsyncMock, patch

import pytest

from app.modules.planning.services.local_coordinator_service import (
    LocalCoordinatorService,
)
from services.coordinator.models import (
    CreatePlanRequest,
    CreatePlanResponse,
    EvaluatorConfig,
    PlanningParameters,
    PortfolioContextInput,
    PositionInput,
    SecurityInput,
)
from services.evaluator.models import ActionCandidateModel as EvalActionCandidate
from services.evaluator.models import (
    EvaluateSequencesResponse,
    SequenceEvaluationResult,
)
from services.generator.models import ActionCandidateModel as GenActionCandidate
from services.generator.models import SequenceBatch
from services.opportunity.models import ActionCandidateModel as OppActionCandidate
from services.opportunity.models import IdentifyOpportunitiesResponse


@pytest.fixture
def coordinator_service():
    """Create LocalCoordinatorService instance."""
    return LocalCoordinatorService()


@pytest.fixture
def sample_request():
    """Create sample CreatePlanRequest."""
    return CreatePlanRequest(
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
        available_cash=5000.0,
        planning_parameters=PlanningParameters(
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
            beam_width=10,
            max_sequence_depth=3,
            batch_size=500,
        ),
        evaluator_config=EvaluatorConfig(
            urls=[
                "http://localhost:8010",
                "http://localhost:8020",
                "http://localhost:8030",
            ]
        ),
    )


@pytest.mark.asyncio
async def test_create_plan_full_workflow(coordinator_service, sample_request):
    """Test full plan creation workflow with all services."""
    # Mock Opportunity Service response
    mock_opportunities = IdentifyOpportunitiesResponse(
        profit_taking=[
            OppActionCandidate(
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
        opportunity_buys=[],
    )

    # Mock Generator Service response (streaming)
    async def mock_generator_stream(*args, **kwargs):
        yield SequenceBatch(
            batch_number=0,
            sequences=[
                [
                    GenActionCandidate(
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
                ]
            ],
            total_batches=1,
            more_available=False,
        )

    # Mock Evaluator Service response
    mock_eval_response = EvaluateSequencesResponse(
        top_sequences=[
            SequenceEvaluationResult(
                sequence=[
                    EvalActionCandidate(
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
                end_state_score=0.85,
                diversification_score=0.75,
                risk_score=0.80,
                total_score=0.85,
                total_cost=5.0,
                cash_required=0.0,  # Sell generates cash
                feasible=True,
                metrics={},
            )
        ],
        total_evaluated=1,
        beam_width=10,
    )

    # Patch HTTP clients
    with (
        patch(
            "app.infrastructure.http_clients.opportunity_client.OpportunityHTTPClient.identify_opportunities",
            new=AsyncMock(return_value=mock_opportunities),
        ),
        patch(
            "app.infrastructure.http_clients.generator_client.GeneratorHTTPClient.generate_sequences_streaming",
            return_value=mock_generator_stream(),
        ),
        patch(
            "app.infrastructure.http_clients.evaluator_client.EvaluatorHTTPClient.evaluate_sequences",
            new=AsyncMock(return_value=mock_eval_response),
        ),
    ):
        response = await coordinator_service.create_plan(sample_request)

    assert isinstance(response, CreatePlanResponse)
    assert len(response.plan.steps) > 0
    assert response.stats.sequences_evaluated == 1
    assert response.stats.batches_processed == 1


@pytest.mark.asyncio
async def test_create_plan_round_robin_distribution(
    coordinator_service, sample_request
):
    """Test that Coordinator distributes batches round-robin across evaluators."""
    # Mock Opportunity Service
    mock_opportunities = IdentifyOpportunitiesResponse(
        profit_taking=[],
        averaging_down=[],
        rebalance_sells=[],
        rebalance_buys=[],
        opportunity_buys=[
            OppActionCandidate(
                side="BUY",
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
        ],
    )

    # Mock Generator Service - 3 batches
    async def mock_generator_stream(*args, **kwargs):
        for i in range(3):
            yield SequenceBatch(
                batch_number=i,
                sequences=[
                    [
                        GenActionCandidate(
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
                ],
                total_batches=3,
                more_available=(i < 2),
            )

    # Mock Evaluator Service
    mock_eval_response = EvaluateSequencesResponse(
        top_sequences=[
            SequenceEvaluationResult(
                sequence=[
                    EvalActionCandidate(
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
                ],
                end_state_score=0.75,
                diversification_score=0.70,
                risk_score=0.80,
                total_score=0.75,
                total_cost=1.2,
                cash_required=201.2,
                feasible=True,
                metrics={},
            )
        ],
        total_evaluated=1,
        beam_width=10,
    )

    evaluator_call_count = [0, 0, 0]  # Track calls to each evaluator

    async def mock_evaluate(self, *args, **kwargs):
        # Determine which evaluator was called based on base_url
        if "8010" in self.base_url:
            evaluator_call_count[0] += 1
        elif "8020" in self.base_url:
            evaluator_call_count[1] += 1
        elif "8030" in self.base_url:
            evaluator_call_count[2] += 1
        return mock_eval_response

    with (
        patch(
            "app.infrastructure.http_clients.opportunity_client.OpportunityHTTPClient.identify_opportunities",
            new=AsyncMock(return_value=mock_opportunities),
        ),
        patch(
            "app.infrastructure.http_clients.generator_client.GeneratorHTTPClient.generate_sequences_streaming",
            return_value=mock_generator_stream(),
        ),
        patch(
            "app.infrastructure.http_clients.evaluator_client.EvaluatorHTTPClient.evaluate_sequences",
            mock_evaluate,
        ),
    ):
        response = await coordinator_service.create_plan(sample_request)

    # Verify 3 batches were processed
    assert response.stats.batches_processed == 3

    # Verify round-robin distribution (each evaluator called once)
    assert evaluator_call_count == [1, 1, 1]


def test_get_next_evaluator_round_robin(coordinator_service):
    """Test round-robin evaluator selection."""
    from app.infrastructure.http_clients.evaluator_client import EvaluatorHTTPClient

    evaluators = [
        EvaluatorHTTPClient(
            base_url="http://localhost:8010", service_name="evaluator-1"
        ),
        EvaluatorHTTPClient(
            base_url="http://localhost:8020", service_name="evaluator-2"
        ),
        EvaluatorHTTPClient(
            base_url="http://localhost:8030", service_name="evaluator-3"
        ),
    ]

    # Should cycle through evaluators in order
    assert (
        coordinator_service._get_next_evaluator(evaluators).base_url
        == "http://localhost:8010"
    )
    assert (
        coordinator_service._get_next_evaluator(evaluators).base_url
        == "http://localhost:8020"
    )
    assert (
        coordinator_service._get_next_evaluator(evaluators).base_url
        == "http://localhost:8030"
    )
    # Should wrap around
    assert (
        coordinator_service._get_next_evaluator(evaluators).base_url
        == "http://localhost:8010"
    )
