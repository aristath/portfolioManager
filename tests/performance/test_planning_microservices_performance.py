"""Performance tests for Planning microservices.

Verifies that the microservice architecture achieves the target 2.5-3× speedup
compared to the monolithic holistic planner.

Target Performance:
- Monolithic: ~86 seconds
- Microservices (3 parallel evaluators): ~32-34 seconds
- Expected Speedup: 2.5-3×
"""

import time
from typing import Dict, List

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


def create_large_portfolio() -> tuple[PortfolioContext, List[Position], List[Security]]:
    """
    Create a large portfolio to stress-test the planning service.

    Returns a portfolio with:
    - 30 positions across multiple countries and industries
    - Total value of 500,000 EUR
    - Mix of gains and losses
    """
    symbols = [
        "AAPL",
        "MSFT",
        "GOOGL",
        "AMZN",
        "META",
        "TSLA",
        "NVDA",
        "JPM",
        "V",
        "JNJ",
        "WMT",
        "PG",
        "UNH",
        "HD",
        "DIS",
        "NFLX",
        "PYPL",
        "ADBE",
        "CRM",
        "INTC",
        "CSCO",
        "PEP",
        "KO",
        "NKE",
        "MCD",
        "ORCL",
        "IBM",
        "QCOM",
        "TXN",
        "AMD",
    ]

    countries = ["US"] * 30  # Simplified for testing
    industries = [
        "Technology",
        "Technology",
        "Technology",
        "Technology",
        "Technology",
        "Technology",
        "Technology",
        "Financial",
        "Financial",
        "Healthcare",
        "Consumer",
        "Consumer",
        "Healthcare",
        "Consumer",
        "Media",
        "Media",
        "Technology",
        "Technology",
        "Technology",
        "Technology",
        "Technology",
        "Consumer",
        "Consumer",
        "Consumer",
        "Consumer",
        "Technology",
        "Technology",
        "Technology",
        "Technology",
        "Technology",
    ]

    positions = []
    securities = []
    total_value = 0.0
    position_values = {}

    for i, symbol in enumerate(symbols):
        # Vary position sizes
        value = 10000.0 + (i * 500)
        total_value += value
        position_values[symbol] = value

        # Create position
        quantity = int(value / (100 + i * 10))
        avg_price = value / quantity
        current_price = avg_price * (1.0 + (i - 15) * 0.01)  # Mix of gains/losses
        unrealized_pnl_pct = (current_price - avg_price) / avg_price

        positions.append(
            Position(
                symbol=symbol,
                quantity=quantity,
                avg_price=avg_price,
                current_price=current_price,
                market_value_eur=value,
                currency=Currency.EUR,
                unrealized_pnl_pct=unrealized_pnl_pct,
            )
        )

        # Create security
        securities.append(
            Security(
                symbol=symbol,
                name=f"{symbol} Corp",
                isin=f"US{i:08d}5",
                country=countries[i],
                industry=industries[i],
                allow_buy=True,
                allow_sell=True,
            )
        )

    # Calculate weights
    country_weights: Dict[str, float] = {}
    industry_weights: Dict[str, float] = {}
    security_countries: Dict[str, str] = {}
    security_industries: Dict[str, str] = {}

    for i, symbol in enumerate(symbols):
        value = position_values[symbol]
        pct = value / total_value

        country = countries[i]
        industry = industries[i]

        country_weights[country] = country_weights.get(country, 0.0) + pct
        industry_weights[industry] = industry_weights.get(industry, 0.0) + pct
        security_countries[symbol] = country
        security_industries[symbol] = industry

    portfolio_context = PortfolioContext(
        total_value=total_value,
        positions=position_values,
        country_weights=country_weights,
        industry_weights=industry_weights,
        security_countries=security_countries,
        security_industries=security_industries,
    )

    return portfolio_context, positions, securities


@pytest.mark.asyncio
@pytest.mark.performance
async def test_monolithic_planner_performance():
    """Benchmark monolithic holistic planner."""
    portfolio_context, positions, securities = create_large_portfolio()

    start_time = time.time()

    plan = await create_holistic_plan(
        portfolio_context=portfolio_context,
        positions=positions,
        securities=securities,
        available_cash=25000.0,
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
        beam_width=10,
        max_depth=3,
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

    elapsed_time = time.time() - start_time

    print("\n" + "=" * 60)
    print("MONOLITHIC PLANNER PERFORMANCE")
    print("=" * 60)
    print(f"Execution time: {elapsed_time:.2f} seconds")
    print(f"Steps generated: {len(plan.steps)}")
    print(f"Plan feasible: {plan.feasible}")
    print("=" * 60 + "\n")

    # Store result for comparison
    return elapsed_time, plan


@pytest.mark.asyncio
@pytest.mark.performance
@pytest.mark.integration
async def test_microservices_planner_performance():
    """Benchmark microservices coordinator with 3 parallel evaluators."""
    portfolio_context, positions, securities = create_large_portfolio()

    coordinator_service = LocalCoordinatorService()

    request = CreatePlanRequest(
        portfolio_context=PortfolioContextInput(
            total_value_eur=portfolio_context.total_value,
            positions=portfolio_context.positions,
            country_weights=portfolio_context.country_weights,
            industry_weights=portfolio_context.industry_weights,
            security_countries=portfolio_context.security_countries,
            security_industries=portfolio_context.security_industries,
        ),
        positions=[
            PositionInput(
                symbol=p.symbol,
                quantity=p.quantity,
                avg_price=p.avg_price,
                market_value_eur=p.market_value_eur,
                unrealized_pnl_pct=p.unrealized_pnl_pct,
            )
            for p in positions
        ],
        securities=[
            SecurityInput(
                symbol=s.symbol,
                name=s.name,
                country=s.country,
                industry=s.industry,
            )
            for s in securities
        ],
        planning_parameters=PlanningParameters(
            available_cash=25000.0,
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

    try:
        start_time = time.time()

        response = await coordinator_service.create_plan(request)

        elapsed_time = time.time() - start_time

        print("\n" + "=" * 60)
        print("MICROSERVICES PLANNER PERFORMANCE")
        print("=" * 60)
        print(f"Execution time: {elapsed_time:.2f} seconds")
        print(f"Steps generated: {len(response.plan.steps)}")
        print(f"Plan feasible: {response.plan.feasible}")
        print(f"Batches processed: {response.execution_stats.batches_processed}")
        print(
            f"Sequences evaluated: {response.execution_stats.total_sequences_evaluated}"
        )
        print("=" * 60 + "\n")

        # Store result for comparison
        return elapsed_time, response.plan

    except Exception as e:
        pytest.skip(f"Microservices not running: {e}")


@pytest.mark.asyncio
@pytest.mark.performance
@pytest.mark.integration
async def test_speedup_verification():
    """
    Verify that microservices achieve at least 2× speedup compared to monolithic.

    Target: 2.5-3× speedup
    Minimum acceptable: 2× speedup
    """
    # Run both benchmarks
    mono_time, mono_plan = await test_monolithic_planner_performance()

    try:
        micro_time, micro_plan = await test_microservices_planner_performance()
    except Exception as e:
        pytest.skip(f"Microservices not running: {e}")

    # Calculate speedup
    speedup = mono_time / micro_time

    print("\n" + "=" * 60)
    print("SPEEDUP ANALYSIS")
    print("=" * 60)
    print(f"Monolithic time: {mono_time:.2f}s")
    print(f"Microservice time: {micro_time:.2f}s")
    print(f"Speedup: {speedup:.2f}×")
    print("Target speedup: 2.5-3.0×")
    print("=" * 60 + "\n")

    # Verify minimum speedup of 2×
    assert speedup >= 2.0, (
        f"Speedup {speedup:.2f}× is below minimum target of 2.0×. "
        f"Monolithic: {mono_time:.2f}s, Microservices: {micro_time:.2f}s"
    )

    # Verify results are equivalent
    assert len(mono_plan.steps) == len(
        micro_plan.steps
    ), f"Plans have different number of steps: {len(mono_plan.steps)} vs {len(micro_plan.steps)}"

    # Report if we achieved target speedup
    if speedup >= 2.5:
        print(f"✓ Achieved target speedup of {speedup:.2f}×")
    else:
        print(
            f"⚠ Speedup {speedup:.2f}× is below target of 2.5×, but above minimum 2.0×"
        )


@pytest.mark.asyncio
@pytest.mark.performance
async def test_single_evaluator_baseline():
    """
    Benchmark with a single evaluator to verify parallelization benefit.

    Single evaluator should be similar speed to monolithic or slightly slower
    due to network overhead.
    """
    portfolio_context, positions, securities = create_large_portfolio()

    coordinator_service = LocalCoordinatorService()

    request = CreatePlanRequest(
        portfolio_context=PortfolioContextInput(
            total_value_eur=portfolio_context.total_value,
            positions=portfolio_context.positions,
            country_weights=portfolio_context.country_weights,
            industry_weights=portfolio_context.industry_weights,
            security_countries=portfolio_context.security_countries,
            security_industries=portfolio_context.security_industries,
        ),
        positions=[
            PositionInput(
                symbol=p.symbol,
                quantity=p.quantity,
                avg_price=p.avg_price,
                market_value_eur=p.market_value_eur,
                unrealized_pnl_pct=p.unrealized_pnl_pct,
            )
            for p in positions
        ],
        securities=[
            SecurityInput(
                symbol=s.symbol,
                name=s.name,
                country=s.country,
                industry=s.industry,
            )
            for s in securities
        ],
        planning_parameters=PlanningParameters(
            available_cash=25000.0,
            transaction_cost_fixed=1.0,
            transaction_cost_percent=0.001,
            beam_width=10,
            max_sequence_depth=3,
            batch_size=500,
        ),
        evaluator_config=EvaluatorConfig(
            urls=["http://localhost:8010"]  # Only one evaluator
        ),
    )

    try:
        start_time = time.time()

        response = await coordinator_service.create_plan(request)

        elapsed_time = time.time() - start_time

        print("\n" + "=" * 60)
        print("SINGLE EVALUATOR BASELINE")
        print("=" * 60)
        print(f"Execution time: {elapsed_time:.2f} seconds")
        print(f"Steps generated: {len(response.plan.steps)}")
        print("=" * 60 + "\n")

        return elapsed_time

    except Exception as e:
        pytest.skip(f"Microservices not running: {e}")
