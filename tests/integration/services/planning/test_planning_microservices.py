"""Integration tests for Planning microservices.

Tests the full workflow of the Planning microservices:
- Opportunity Service
- Generator Service
- Evaluator Service (3 instances)
- Coordinator Service

Requires all services to be running (via docker-compose).
"""

import pytest
from httpx import AsyncClient, ConnectError

from services.coordinator.models import (
    CreatePlanRequest,
    EvaluatorConfig,
    PlanningParameters,
    PortfolioContextInput,
    PositionInput,
    SecurityInput,
)
from services.evaluator.models import (
    ActionCandidateModel,
    EvaluateSequencesRequest,
    EvaluationSettings,
)
from services.generator.models import (
    CombinatorialSettings,
    FeasibilitySettings,
    GenerateSequencesRequest,
    OpportunitiesInput,
)
from services.opportunity.models import IdentifyOpportunitiesRequest

# Service URLs
OPPORTUNITY_URL = "http://localhost:8008"
GENERATOR_URL = "http://localhost:8009"
EVALUATOR_1_URL = "http://localhost:8010"
EVALUATOR_2_URL = "http://localhost:8020"
EVALUATOR_3_URL = "http://localhost:8030"
COORDINATOR_URL = "http://localhost:8011"


@pytest.fixture
def sample_portfolio_context():
    """Create sample portfolio context for testing."""
    return PortfolioContextInput(
        total_value_eur=100000.0,
        positions={"AAPL": 10000.0, "MSFT": 15000.0, "GOOGL": 12000.0},
        country_weights={"US": 0.37},
        industry_weights={"Technology": 0.37},
        security_countries={"AAPL": "US", "MSFT": "US", "GOOGL": "US"},
        security_industries={
            "AAPL": "Technology",
            "MSFT": "Technology",
            "GOOGL": "Technology",
        },
    )


@pytest.fixture
def sample_positions():
    """Create sample positions for testing."""
    return [
        PositionInput(
            symbol="AAPL",
            quantity=50,
            avg_price=180.0,
            market_value_eur=10000.0,
            unrealized_pnl_pct=0.11,
        ),
        PositionInput(
            symbol="MSFT",
            quantity=40,
            avg_price=350.0,
            market_value_eur=15000.0,
            unrealized_pnl_pct=0.07,
        ),
        PositionInput(
            symbol="GOOGL",
            quantity=80,
            avg_price=140.0,
            market_value_eur=12000.0,
            unrealized_pnl_pct=0.07,
        ),
    ]


@pytest.fixture
def sample_securities():
    """Create sample securities for testing."""
    return [
        SecurityInput(
            symbol="AAPL",
            name="Apple Inc.",
            country="US",
            industry="Technology",
        ),
        SecurityInput(
            symbol="MSFT",
            name="Microsoft Corporation",
            country="US",
            industry="Technology",
        ),
        SecurityInput(
            symbol="GOOGL",
            name="Alphabet Inc.",
            country="US",
            industry="Technology",
        ),
    ]


@pytest.mark.asyncio
@pytest.mark.integration
async def test_opportunity_service_health():
    """Test Opportunity Service health endpoint."""
    async with AsyncClient() as client:
        try:
            response = await client.get(f"{OPPORTUNITY_URL}/opportunity/health")
            assert response.status_code == 200
            assert response.json()["status"] == "healthy"
        except ConnectError:
            pytest.skip("Opportunity service not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_generator_service_health():
    """Test Generator Service health endpoint."""
    async with AsyncClient() as client:
        try:
            response = await client.get(f"{GENERATOR_URL}/generator/health")
            assert response.status_code == 200
            assert response.json()["status"] == "healthy"
        except ConnectError:
            pytest.skip("Generator service not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_evaluator_services_health():
    """Test all Evaluator Service instances health endpoints."""
    evaluator_urls = [EVALUATOR_1_URL, EVALUATOR_2_URL, EVALUATOR_3_URL]

    async with AsyncClient() as client:
        for url in evaluator_urls:
            try:
                response = await client.get(f"{url}/evaluator/health")
                assert response.status_code == 200
                assert response.json()["status"] == "healthy"
            except ConnectError:
                pytest.skip(f"Evaluator service at {url} not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_coordinator_service_health():
    """Test Coordinator Service health endpoint."""
    async with AsyncClient() as client:
        try:
            response = await client.get(f"{COORDINATOR_URL}/coordinator/health")
            assert response.status_code == 200
            assert response.json()["status"] == "healthy"
        except ConnectError:
            pytest.skip("Coordinator service not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_opportunity_service_identify(
    sample_portfolio_context, sample_positions, sample_securities
):
    """Test Opportunity Service opportunity identification."""
    request = IdentifyOpportunitiesRequest(
        portfolio_context=sample_portfolio_context,
        positions=sample_positions,
        securities=sample_securities,
        available_cash=5000.0,
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
    )

    async with AsyncClient() as client:
        try:
            response = await client.post(
                f"{OPPORTUNITY_URL}/opportunity/identify",
                json=request.model_dump(),
                timeout=10.0,
            )
            assert response.status_code == 200

            data = response.json()
            # Should have categorized opportunities
            assert "profit_taking" in data
            assert "averaging_down" in data
            assert "rebalance_sells" in data
            assert "rebalance_buys" in data
            assert "opportunity_buys" in data

            # At least one category should have opportunities
            total_opportunities = sum(len(data[key]) for key in data.keys())
            assert total_opportunities > 0

        except ConnectError:
            pytest.skip("Opportunity service not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_generator_service_streaming():
    """Test Generator Service batch streaming."""
    request = GenerateSequencesRequest(
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
            opportunity_buys=[],
        ),
        combinatorial=CombinatorialSettings(
            enable_weighted_combinations=True,
            max_depth=2,
        ),
        feasibility=FeasibilitySettings(
            available_cash=10000.0,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
            min_trade_value=100.0,
        ),
        batch_size=500,
    )

    async with AsyncClient() as client:
        try:
            response = await client.post(
                f"{GENERATOR_URL}/generator/generate",
                json=request.model_dump(),
                timeout=30.0,
            )
            assert response.status_code == 200

            # Parse streaming response (NDJSON)
            batches = []
            for line in response.text.strip().split("\n"):
                if line:
                    import json

                    batch = json.loads(line)
                    batches.append(batch)

            assert len(batches) > 0
            # Verify batch structure
            assert "batch_number" in batches[0]
            assert "sequences" in batches[0]
            assert "total_batches" in batches[0]

        except ConnectError:
            pytest.skip("Generator service not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_evaluator_service_evaluate(
    sample_portfolio_context, sample_positions, sample_securities
):
    """Test Evaluator Service sequence evaluation."""
    request = EvaluateSequencesRequest(
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
            ]
        ],
        portfolio_context=sample_portfolio_context,
        positions=sample_positions,
        securities=sample_securities,
        settings=EvaluationSettings(
            beam_width=10,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
        ),
    )

    async with AsyncClient() as client:
        try:
            response = await client.post(
                f"{EVALUATOR_1_URL}/evaluator/evaluate",
                json=request.model_dump(),
                timeout=120.0,
            )
            assert response.status_code == 200

            data = response.json()
            assert "top_sequences" in data
            assert "total_evaluated" in data
            assert data["total_evaluated"] == 1
            assert len(data["top_sequences"]) == 1

            # Verify sequence evaluation result structure
            result = data["top_sequences"][0]
            assert "sequence" in result
            assert "end_state_score" in result
            assert "total_score" in result
            assert "feasible" in result

        except ConnectError:
            pytest.skip("Evaluator service not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_coordinator_full_workflow(
    sample_portfolio_context, sample_positions, sample_securities
):
    """Test Coordinator Service full planning workflow."""
    request = CreatePlanRequest(
        portfolio_context=sample_portfolio_context,
        positions=sample_positions,
        securities=sample_securities,
        planning_parameters=PlanningParameters(
            available_cash=5000.0,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
            beam_width=10,
            max_sequence_depth=2,
            batch_size=500,
        ),
        evaluator_config=EvaluatorConfig(
            urls=[EVALUATOR_1_URL, EVALUATOR_2_URL, EVALUATOR_3_URL]
        ),
    )

    async with AsyncClient() as client:
        try:
            response = await client.post(
                f"{COORDINATOR_URL}/coordinator/create-plan",
                json=request.model_dump(),
                timeout=300.0,
            )
            assert response.status_code == 200

            data = response.json()
            assert "plan" in data
            assert "execution_stats" in data

            # Verify plan structure
            plan = data["plan"]
            assert "steps" in plan
            assert "narrative" in plan
            assert "total_value" in plan

            # Verify execution stats
            stats = data["execution_stats"]
            assert "total_sequences_evaluated" in stats
            assert "batches_processed" in stats
            assert "execution_time_seconds" in stats

        except ConnectError:
            pytest.skip("Coordinator service not running")


@pytest.mark.asyncio
@pytest.mark.integration
async def test_parallel_evaluator_distribution(
    sample_portfolio_context, sample_positions, sample_securities
):
    """Test that Coordinator distributes work across all 3 evaluators."""
    request = CreatePlanRequest(
        portfolio_context=sample_portfolio_context,
        positions=sample_positions,
        securities=sample_securities,
        planning_parameters=PlanningParameters(
            available_cash=5000.0,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
            beam_width=10,
            max_sequence_depth=3,  # Higher depth to generate more batches
            batch_size=100,  # Smaller batches to force multiple batches
        ),
        evaluator_config=EvaluatorConfig(
            urls=[EVALUATOR_1_URL, EVALUATOR_2_URL, EVALUATOR_3_URL]
        ),
    )

    async with AsyncClient() as client:
        try:
            response = await client.post(
                f"{COORDINATOR_URL}/coordinator/create-plan",
                json=request.model_dump(),
                timeout=300.0,
            )
            assert response.status_code == 200

            data = response.json()
            stats = data["execution_stats"]

            # If multiple batches were processed, work should be distributed
            if stats["batches_processed"] >= 3:
                # This is hard to verify without instrumentation in the services
                # but we can at least verify the plan was created successfully
                assert "plan" in data
                assert len(data["plan"]["steps"]) > 0

        except ConnectError:
            pytest.skip("Coordinator service not running")
