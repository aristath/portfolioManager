"""Unit tests for LocalOpportunityService."""

from unittest.mock import AsyncMock, patch

import pytest

from app.domain.models import Position, Security
from app.modules.planning.domain.models import ActionCandidate
from app.modules.planning.services.local_opportunity_service import (
    LocalOpportunityService,
)
from app.modules.scoring.domain.models import PortfolioContext
from app.shared.domain.value_objects.currency import Currency
from services.opportunity.models import (
    IdentifyOpportunitiesRequest,
    IdentifyOpportunitiesResponse,
    PortfolioContextInput,
    PositionInput,
    SecurityInput,
)


@pytest.fixture
def opportunity_service():
    """Create LocalOpportunityService instance."""
    return LocalOpportunityService()


@pytest.fixture
def sample_request():
    """Create sample IdentifyOpportunitiesRequest."""
    return IdentifyOpportunitiesRequest(
        portfolio_context=PortfolioContextInput(
            total_value=100000.0,
            positions={"AAPL": 10000.0, "MSFT": 15000.0},
            country_weights={"US": 0.25},
            industry_weights={"Technology": 0.25},
            security_countries={"AAPL": "US", "MSFT": "US"},
            security_industries={"AAPL": "Technology", "MSFT": "Technology"},
        ),
        positions=[
            PositionInput(
                symbol="AAPL",
                quantity=50,
                avg_price=180.0,
                market_value_eur=10000.0,
                unrealized_pnl_pct=0.1,
            ),
            PositionInput(
                symbol="MSFT",
                quantity=40,
                avg_price=350.0,
                market_value_eur=15000.0,
                unrealized_pnl_pct=0.15,
            ),
        ],
        securities=[
            SecurityInput(
                symbol="AAPL",
                name="Apple Inc.",
                isin="US0378331005",
                country="US",
                industry="Technology",
                allow_buy=True,
                allow_sell=True,
            ),
            SecurityInput(
                symbol="MSFT",
                name="Microsoft Corporation",
                isin="US5949181045",
                country="US",
                industry="Technology",
                allow_buy=True,
                allow_sell=True,
            ),
        ],
        available_cash=5000.0,
        transaction_cost_fixed=1.0,
        transaction_cost_percent=0.001,
    )


@pytest.mark.asyncio
async def test_identify_opportunities_without_weights(
    opportunity_service, sample_request
):
    """Test opportunity identification without target weights (heuristic mode)."""
    # Mock the domain function
    mock_opportunities = {
        "profit_taking": [
            ActionCandidate(
                side="SELL",
                symbol="MSFT",
                name="Microsoft Corporation",
                quantity=10,
                price=400.0,
                value_eur=4000.0,
                currency="EUR",
                priority=1,
                reason="High unrealized gains",
                tags=["profit_taking"],
            )
        ],
        "averaging_down": [],
        "rebalance_sells": [],
        "rebalance_buys": [],
        "opportunity_buys": [],
    }

    with patch(
        "app.modules.planning.services.local_opportunity_service.identify_opportunities",
        new=AsyncMock(return_value=mock_opportunities),
    ):
        response = await opportunity_service.identify_opportunities(sample_request)

    assert isinstance(response, IdentifyOpportunitiesResponse)
    assert len(response.profit_taking) == 1
    assert response.profit_taking[0].symbol == "MSFT"
    assert response.profit_taking[0].side == "SELL"
    assert len(response.averaging_down) == 0


@pytest.mark.asyncio
async def test_identify_opportunities_with_weights(opportunity_service, sample_request):
    """Test opportunity identification with target weights (weight-based mode)."""
    # Add target weights to request
    sample_request.target_weights = {"AAPL": 0.15, "MSFT": 0.10}

    mock_opportunities = {
        "profit_taking": [],
        "averaging_down": [],
        "rebalance_sells": [
            ActionCandidate(
                side="SELL",
                symbol="MSFT",
                name="Microsoft Corporation",
                quantity=5,
                price=400.0,
                value_eur=2000.0,
                currency="EUR",
                priority=1,
                reason="Overweight position",
                tags=["rebalance"],
            )
        ],
        "rebalance_buys": [],
        "opportunity_buys": [],
    }

    with patch(
        "app.modules.planning.services.local_opportunity_service.identify_opportunities_from_weights",
        new=AsyncMock(return_value=mock_opportunities),
    ):
        response = await opportunity_service.identify_opportunities(sample_request)

    assert isinstance(response, IdentifyOpportunitiesResponse)
    assert len(response.rebalance_sells) == 1
    assert response.rebalance_sells[0].symbol == "MSFT"


def test_to_portfolio_context(opportunity_service, sample_request):
    """Test conversion from request to PortfolioContext."""
    context = opportunity_service._to_portfolio_context(sample_request)

    assert isinstance(context, PortfolioContext)
    assert context.total_value == 100000.0
    assert context.positions == {"AAPL": 10000.0, "MSFT": 15000.0}
    assert context.country_weights == {"US": 0.25}


def test_to_positions(opportunity_service, sample_request):
    """Test conversion from PositionInput to Position domain models."""
    positions = opportunity_service._to_positions(sample_request.positions)

    assert len(positions) == 2
    assert all(isinstance(p, Position) for p in positions)
    assert positions[0].symbol == "AAPL"
    assert positions[0].quantity == 50
    assert positions[0].avg_price == 180.0
    assert positions[0].currency == Currency.EUR


def test_to_securities(opportunity_service, sample_request):
    """Test conversion from SecurityInput to Security domain models."""
    securities = opportunity_service._to_securities(sample_request.securities)

    assert len(securities) == 2
    assert all(isinstance(s, Security) for s in securities)
    assert securities[0].symbol == "AAPL"
    assert securities[0].name == "Apple Inc."
    assert securities[0].country == "US"
    assert securities[0].allow_buy is True


def test_action_candidate_to_model(opportunity_service):
    """Test conversion from ActionCandidate to ActionCandidateModel."""
    from app.domain.value_objects.trade_side import TradeSide

    action = ActionCandidate(
        side=TradeSide.BUY,
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

    model = opportunity_service._action_candidate_to_model(action)

    assert model.side == "BUY"
    assert model.symbol == "AAPL"
    assert model.quantity == 10
    assert model.value_eur == 2000.0
    assert model.tags == ["opportunity_buy"]
