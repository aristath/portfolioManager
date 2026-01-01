"""Equivalence test for Planning microservices vs monolithic planner.

Verifies that the microservice architecture produces the same results
as the monolithic holistic planner.
"""

import pytest

from app.domain.models import Position, Security
from app.modules.planning.domain.holistic_planner import create_holistic_plan
from app.modules.planning.services.local_coordinator_service import (
    LocalCoordinatorService,
)
from app.modules.scoring.domain.models import PortfolioContext
from app.shared.domain.value_objects.currency import Currency
from services.coordinator.models import (
    CreatePlanRequest,
    EvaluatorConfig,
    PlanningParameters,
    PortfolioContextInput,
    PositionInput,
    SecurityInput,
)


@pytest.fixture
def sample_portfolio_context():
    """Create sample portfolio context."""
    return PortfolioContext(
        total_value=100000.0,
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
    """Create sample positions."""
    return [
        Position(
            symbol="AAPL",
            quantity=50,
            avg_price=180.0,
            current_price=200.0,
            market_value_eur=10000.0,
            currency=Currency.EUR,
            unrealized_pnl_pct=0.11,
        ),
        Position(
            symbol="MSFT",
            quantity=40,
            avg_price=350.0,
            current_price=375.0,
            market_value_eur=15000.0,
            currency=Currency.EUR,
            unrealized_pnl_pct=0.07,
        ),
        Position(
            symbol="GOOGL",
            quantity=80,
            avg_price=140.0,
            current_price=150.0,
            market_value_eur=12000.0,
            currency=Currency.EUR,
            unrealized_pnl_pct=0.07,
        ),
    ]


@pytest.fixture
def sample_securities():
    """Create sample securities."""
    return [
        Security(
            symbol="AAPL",
            name="Apple Inc.",
            isin="US0378331005",
            country="US",
            industry="Technology",
            allow_buy=True,
            allow_sell=True,
        ),
        Security(
            symbol="MSFT",
            name="Microsoft Corporation",
            isin="US5949181045",
            country="US",
            industry="Technology",
            allow_buy=True,
            allow_sell=True,
        ),
        Security(
            symbol="GOOGL",
            name="Alphabet Inc.",
            isin="US02079K3059",
            country="US",
            industry="Technology",
            allow_buy=True,
            allow_sell=True,
        ),
    ]


@pytest.mark.asyncio
@pytest.mark.integration
async def test_microservice_vs_monolithic_equivalence(
    sample_portfolio_context, sample_positions, sample_securities
):
    """
    Test that microservice architecture produces equivalent results to monolithic planner.

    This test runs both implementations with the same inputs and verifies:
    1. Same number of steps in the plan
    2. Same actions (symbol, side, quantity)
    3. Similar scores (within tolerance)
    4. Same feasibility determination
    """
    # Common parameters
    available_cash = 5000.0
    transaction_cost_fixed = 1.0
    transaction_cost_percent = 0.001
    beam_width = 10
    max_depth = 2

    # ========== Run Monolithic Planner ==========
    monolithic_plan = await create_holistic_plan(
        portfolio_context=sample_portfolio_context,
        positions=sample_positions,
        securities=sample_securities,
        available_cash=available_cash,
        transaction_cost_fixed=transaction_cost_fixed,
        transaction_cost_percent=transaction_cost_percent,
        beam_width=beam_width,
        max_depth=max_depth,
        target_weights=None,
        current_prices={},
        enable_combinatorial=True,
        enable_correlation_aware=False,
        enable_partial_execution=False,
        enable_constraint_relaxation=False,
        enable_monte_carlo=False,
        enable_stochastic_scenarios=False,
        monte_carlo_iterations=0,
    )

    # ========== Run Microservice Coordinator ==========
    coordinator_service = LocalCoordinatorService()

    # Convert domain models to Pydantic models for microservice request
    request = CreatePlanRequest(
        portfolio_context=PortfolioContextInput(
            total_value_eur=sample_portfolio_context.total_value,
            positions=sample_portfolio_context.positions,
            country_weights=sample_portfolio_context.country_weights,
            industry_weights=sample_portfolio_context.industry_weights,
            security_countries=sample_portfolio_context.security_countries,
            security_industries=sample_portfolio_context.security_industries,
        ),
        positions=[
            PositionInput(
                symbol=p.symbol,
                quantity=p.quantity,
                avg_price=p.avg_price,
                market_value_eur=p.market_value_eur,
                unrealized_pnl_pct=p.unrealized_pnl_pct,
            )
            for p in sample_positions
        ],
        securities=[
            SecurityInput(
                symbol=s.symbol,
                name=s.name,
                country=s.country,
                industry=s.industry,
            )
            for s in sample_securities
        ],
        planning_parameters=PlanningParameters(
            available_cash=available_cash,
            transaction_cost_fixed=transaction_cost_fixed,
            transaction_cost_percent=transaction_cost_percent,
            beam_width=beam_width,
            max_sequence_depth=max_depth,
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

    # This test requires services to be running
    try:
        microservice_response = await coordinator_service.create_plan(request)
    except Exception as e:
        pytest.skip(f"Microservices not running: {e}")

    microservice_plan = microservice_response.plan

    # ========== Compare Results ==========

    # 1. Compare number of steps
    assert len(monolithic_plan.steps) == len(microservice_plan.steps), (
        f"Different number of steps: monolithic={len(monolithic_plan.steps)}, "
        f"microservice={len(microservice_plan.steps)}"
    )

    # 2. Compare each step
    for i, (mono_step, micro_step) in enumerate(
        zip(monolithic_plan.steps, microservice_plan.steps)
    ):
        assert (
            mono_step.symbol == micro_step.symbol
        ), f"Step {i}: Different symbols: {mono_step.symbol} vs {micro_step.symbol}"

        assert (
            mono_step.side == micro_step.side
        ), f"Step {i}: Different sides: {mono_step.side} vs {micro_step.side}"

        # Quantity might differ slightly due to rounding
        quantity_diff = abs(mono_step.quantity - micro_step.quantity)
        assert quantity_diff <= 1, (
            f"Step {i}: Quantities differ by more than 1: "
            f"{mono_step.quantity} vs {micro_step.quantity}"
        )

        # Value EUR should be close
        value_diff_pct = (
            abs(mono_step.value_eur - micro_step.value_eur) / mono_step.value_eur
        )
        assert value_diff_pct < 0.01, (
            f"Step {i}: Values differ by more than 1%: "
            f"{mono_step.value_eur} vs {micro_step.value_eur}"
        )

    # 3. Compare scores (allow small tolerance due to floating point)
    score_tolerance = 0.01

    assert (
        abs(monolithic_plan.end_state_score - microservice_plan.end_state_score)
        < score_tolerance
    ), (
        f"End state scores differ: {monolithic_plan.end_state_score} vs "
        f"{microservice_plan.end_state_score}"
    )

    # 4. Compare feasibility
    assert (
        monolithic_plan.feasible == microservice_plan.feasible
    ), f"Feasibility differs: {monolithic_plan.feasible} vs {microservice_plan.feasible}"

    # 5. Compare cash required (allow small tolerance)
    cash_diff_pct = abs(
        monolithic_plan.cash_required - microservice_plan.cash_required
    ) / max(monolithic_plan.cash_required, 1.0)
    assert cash_diff_pct < 0.01, (
        f"Cash required differs by more than 1%: "
        f"{monolithic_plan.cash_required} vs {microservice_plan.cash_required}"
    )

    # 6. Compare total value
    value_diff_pct = (
        abs(monolithic_plan.total_value - microservice_plan.total_value)
        / monolithic_plan.total_value
    )
    assert value_diff_pct < 0.01, (
        f"Total values differ by more than 1%: "
        f"{monolithic_plan.total_value} vs {microservice_plan.total_value}"
    )


@pytest.mark.asyncio
@pytest.mark.integration
async def test_empty_portfolio_equivalence():
    """Test equivalence with empty portfolio (edge case)."""
    portfolio_context = PortfolioContext(
        total_value=0.0,
        positions={},
        country_weights={},
        industry_weights={},
        security_countries={},
        security_industries={},
    )

    positions = []
    securities = []

    # Run monolithic
    monolithic_plan = await create_holistic_plan(
        portfolio_context=portfolio_context,
        positions=positions,
        securities=securities,
        available_cash=1000.0,
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
        beam_width=10,
        max_depth=2,
        target_weights=None,
        current_prices={},
        enable_combinatorial=True,
        enable_correlation_aware=False,
        enable_partial_execution=False,
        enable_constraint_relaxation=False,
        enable_monte_carlo=False,
        enable_stochastic_scenarios=False,
        monte_carlo_iterations=0,
    )

    # Run microservice
    coordinator_service = LocalCoordinatorService()
    request = CreatePlanRequest(
        portfolio_context=PortfolioContextInput(
            total_value_eur=0.0,
            positions={},
            country_weights={},
            industry_weights={},
            security_countries={},
            security_industries={},
        ),
        positions=[],
        securities=[],
        planning_parameters=PlanningParameters(
            available_cash=1000.0,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
            beam_width=10,
            max_sequence_depth=2,
            batch_size=500,
        ),
        evaluator_config=EvaluatorConfig(urls=["http://localhost:8010"]),
    )

    try:
        microservice_response = await coordinator_service.create_plan(request)
    except Exception as e:
        pytest.skip(f"Microservices not running: {e}")

    microservice_plan = microservice_response.plan

    # Both should produce empty plans
    assert len(monolithic_plan.steps) == len(microservice_plan.steps)
    assert monolithic_plan.feasible == microservice_plan.feasible
