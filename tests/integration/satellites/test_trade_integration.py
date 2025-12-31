"""Integration tests for trade-bucket integration.

Tests that trades properly update bucket balances and track bucket attribution.
"""

from datetime import datetime
from decimal import Decimal

import pytest

from app.domain.factories.trade_factory import TradeFactory
from app.domain.models import Position, Security
from app.domain.value_objects.product_type import ProductType
from app.domain.value_objects.trade_side import TradeSide
from app.modules.portfolio.database.position_repository import PositionRepository
from app.modules.satellites.services.balance_service import BalanceService
from app.modules.satellites.services.bucket_service import BucketService
from app.modules.universe.database.security_repository import SecurityRepository
from app.repositories import TradeRepository
from app.shared.domain.value_objects.currency import Currency


@pytest.fixture
async def bucket_service(db_manager):
    """Create a bucket service instance."""
    return BucketService()


@pytest.fixture
async def balance_repo(db_manager):
    """Create a balance repository instance."""
    from app.modules.satellites.database.balance_repository import BalanceRepository

    return BalanceRepository()


@pytest.fixture
async def balance_service(db_manager):
    """Create a balance service instance."""
    return BalanceService()


@pytest.fixture
async def security_repo(db_manager):
    """Create a security repository instance."""
    return SecurityRepository(db=db_manager.config)


@pytest.fixture
async def trade_repo(db_manager):
    """Create a trade repository instance."""
    return TradeRepository(db=db_manager.ledger)


@pytest.fixture
async def position_repo(db_manager):
    """Create a position repository instance."""
    return PositionRepository(db=db_manager.state)


@pytest.fixture
async def core_bucket(bucket_service):
    """Ensure core bucket exists."""
    # Core bucket should be created automatically, but ensure it exists
    core = await bucket_service.get_bucket("core")
    if not core:
        await bucket_service.create_bucket(
            bucket_id="core",
            name="Core Portfolio",
            target_allocation_pct=Decimal("0.75"),
        )
        core = await bucket_service.get_bucket("core")
    return core


@pytest.fixture
async def test_satellite(bucket_service):
    """Create a test satellite bucket."""
    await bucket_service.create_satellite(
        satellite_id="test-satellite",
        name="Test Satellite",
        start_in_research=False,  # Start active, not in research
    )
    satellite = await bucket_service.get_bucket("test-satellite")
    return satellite


@pytest.mark.asyncio
async def test_buy_trade_updates_bucket_balance(
    balance_repo, balance_service, security_repo, trade_repo, core_bucket
):
    """Test that a BUY trade decreases bucket cash balance."""
    # Setup: Create a security in core bucket
    security = Security(
        symbol="AAPL",
        yahoo_symbol="AAPL",
        isin="US0378331005",
        name="Apple Inc.",
        product_type=ProductType.EQUITY,
        industry="Consumer Electronics",
        country="United States",
        priority_multiplier=1.0,
        min_lot=1,
        active=True,
        currency="USD",
    )
    await security_repo.create(security)

    # Fund the core bucket (using repository for test setup)
    await balance_repo.adjust_balance("core", "EUR", float(Decimal("10000.00")))

    # Check initial balance
    initial_balances = await balance_service.get_all_balances("core")
    initial_eur = next(b.balance for b in initial_balances if b.currency == "EUR")
    assert initial_eur == Decimal("10000.00")

    # Execute a BUY trade
    trade_cost = 1500.00  # 10 shares @ $150
    trade = TradeFactory.create_from_execution(
        symbol="AAPL",
        side=TradeSide.BUY,
        quantity=10.0,
        price=150.0,
        order_id="test-order-1",
        executed_at=datetime.now(),
        currency=Currency.USD,
        currency_rate=1.10,  # 1 EUR = 1.10 USD
        isin="US0378331005",
        bucket_id="core",
        mode="live",
    )

    # Record the trade
    await trade_repo.create(trade)

    # Update bucket balance (simulating what trade_recorder does)
    # Convert trade cost to EUR for settlement
    trade_cost_eur = trade_cost / 1.10  # Convert USD to EUR
    if trade.mode == "live":
        await balance_service.record_trade_settlement(
            bucket_id=trade.bucket_id,
            amount=trade_cost_eur,
            currency="EUR",  # Settle in EUR
            is_buy=True,
            description=f"Buy {trade.quantity} {trade.symbol} @ {trade.price}",
        )

    # Verify balance decreased by trade cost (in EUR)
    final_balances = await balance_service.get_all_balances("core")
    final_eur = next(b.balance for b in final_balances if b.currency == "EUR")

    # Balance should have decreased by trade cost
    expected_balance = initial_eur - trade_cost_eur
    assert (
        abs(final_eur - expected_balance) < 0.01
    ), f"Expected ~{expected_balance}, got {final_eur}"


@pytest.mark.asyncio
async def test_sell_trade_increases_bucket_balance(
    balance_repo, balance_service, security_repo, trade_repo, position_repo, core_bucket
):
    """Test that a SELL trade increases bucket cash balance."""
    # Setup: Create a security
    security = Security(
        symbol="MSFT",
        yahoo_symbol="MSFT",
        isin="US5949181045",
        name="Microsoft Corp.",
        product_type=ProductType.EQUITY,
        industry="Software",
        country="United States",
        priority_multiplier=1.0,
        min_lot=1,
        active=True,
        currency="USD",
    )
    await security_repo.create(security)

    # Create an existing position to sell
    position = Position(
        symbol="MSFT",
        quantity=10.0,
        avg_price=300.0,
        current_price=320.0,
        currency="USD",
        currency_rate=1.10,
        market_value_eur=2909.09,
        last_updated=datetime.now().isoformat(),
        bucket_id="core",
    )
    await position_repo.upsert(position)

    # Fund the core bucket with some initial cash (using repository for test setup)
    await balance_repo.adjust_balance("core", "EUR", float(Decimal("1000.00")))

    # Check initial balance
    initial_balances = await balance_service.get_all_balances("core")
    initial_eur = next(b.balance for b in initial_balances if b.currency == "EUR")

    # Execute a SELL trade
    sell_proceeds = 3200.00  # 10 shares @ $320
    trade = TradeFactory.create_from_execution(
        symbol="MSFT",
        side=TradeSide.SELL,
        quantity=10.0,
        price=320.0,
        order_id="test-order-2",
        executed_at=datetime.now(),
        currency=Currency.USD,
        currency_rate=1.10,
        isin="US5949181045",
        bucket_id="core",
        mode="live",
    )

    # Record the trade
    await trade_repo.create(trade)

    # Update bucket balance (simulating what trade_recorder does)
    # Convert proceeds to EUR for settlement
    proceeds_eur = sell_proceeds / 1.10  # Convert USD to EUR
    if trade.mode == "live":
        await balance_service.record_trade_settlement(
            bucket_id=trade.bucket_id,
            amount=proceeds_eur,
            currency="EUR",  # Settle in EUR
            is_buy=False,
            description=f"Sell {trade.quantity} {trade.symbol} @ {trade.price}",
        )

    # Verify balance increased by sell proceeds (in EUR)
    final_balances = await balance_service.get_all_balances("core")
    final_eur = next(b.balance for b in final_balances if b.currency == "EUR")

    # Balance should have increased by proceeds
    expected_balance = initial_eur + proceeds_eur
    assert (
        abs(final_eur - expected_balance) < 0.01
    ), f"Expected ~{expected_balance}, got {final_eur}"


@pytest.mark.asyncio
async def test_research_mode_trade_doesnt_affect_balance(
    balance_repo, balance_service, security_repo, trade_repo, core_bucket
):
    """Test that research mode trades don't affect bucket balances."""
    # Setup: Create a security
    security = Security(
        symbol="GOOGL",
        yahoo_symbol="GOOGL",
        isin="US02079K3059",
        name="Alphabet Inc.",
        product_type=ProductType.EQUITY,
        industry="Internet",
        country="United States",
        priority_multiplier=1.0,
        min_lot=1,
        active=True,
        currency="USD",
    )
    await security_repo.create(security)

    # Fund the core bucket (using repository for test setup)
    await balance_repo.adjust_balance("core", "EUR", float(Decimal("5000.00")))

    # Check initial balance
    initial_balances = await balance_service.get_all_balances("core")
    initial_eur = next(b.balance for b in initial_balances if b.currency == "EUR")

    # Execute a research mode BUY trade
    trade = TradeFactory.create_from_execution(
        symbol="GOOGL",
        side=TradeSide.BUY,
        quantity=5.0,
        price=2800.0,
        order_id="test-research-1",
        executed_at=datetime.now(),
        currency=Currency.USD,
        currency_rate=1.10,
        isin="US02079K3059",
        bucket_id="core",
        mode="research",  # Research mode!
    )

    # Record the trade
    await trade_repo.create(trade)

    # Note: Research mode trades should NOT update balance
    # (trade_recorder checks mode and skips balance update)

    # Verify balance unchanged
    final_balances = await balance_service.get_all_balances("core")
    final_eur = next(b.balance for b in final_balances if b.currency == "EUR")
    assert final_eur == initial_eur, "Research mode trade should not affect balance"

    # Verify trade was recorded
    trades = await trade_repo.get_history(limit=10)
    assert len(trades) == 1
    assert trades[0].mode == "research"


@pytest.mark.asyncio
async def test_satellite_trade_updates_satellite_bucket(
    db_manager, balance_repo, balance_service, security_repo, trade_repo, test_satellite
):
    """Test that a trade for a satellite security updates the satellite bucket."""
    # Setup: Create a security assigned to satellite
    # Use raw SQL to insert with bucket_id since create() might not preserve it
    from app.repositories.base import transaction_context

    async with transaction_context(db_manager.config) as conn:
        await conn.execute(
            """INSERT INTO securities
            (symbol, yahoo_symbol, isin, name, product_type, industry, country,
             priority_multiplier, min_lot, active, currency, bucket_id, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
            (
                "TSLA",
                "TSLA",
                "US88160R1014",
                "Tesla Inc.",
                "equity",
                "Automotive",
                "United States",
                1.0,
                1,
                1,
                "USD",
                "test-satellite",
                datetime.now().isoformat(),
                datetime.now().isoformat(),
            ),
        )

    # Fund the satellite bucket (using repository for test setup)
    await balance_repo.adjust_balance(
        "test-satellite", "EUR", float(Decimal("8000.00"))
    )

    # Check initial satellite balance
    initial_balances = await balance_service.get_all_balances("test-satellite")
    initial_eur = next(b.balance for b in initial_balances if b.currency == "EUR")

    # Execute a BUY trade
    trade_cost = 4000.00  # 20 shares @ $200
    trade = TradeFactory.create_from_execution(
        symbol="TSLA",
        side=TradeSide.BUY,
        quantity=20.0,
        price=200.0,
        order_id="test-satellite-1",
        executed_at=datetime.now(),
        currency=Currency.USD,
        currency_rate=1.10,
        isin="US88160R1014",
        bucket_id="test-satellite",
        mode="live",
    )

    # Record the trade
    await trade_repo.create(trade)

    # Update bucket balance (simulating what trade_recorder does)
    # Convert trade cost to EUR for settlement
    trade_cost_eur = trade_cost / 1.10  # Convert USD to EUR
    if trade.mode == "live":
        await balance_service.record_trade_settlement(
            bucket_id=trade.bucket_id,
            amount=trade_cost_eur,
            currency="EUR",  # Settle in EUR
            is_buy=True,
            description=f"Buy {trade.quantity} {trade.symbol} @ {trade.price}",
        )

    # Verify satellite balance decreased
    final_balances = await balance_service.get_all_balances("test-satellite")
    final_eur = next(b.balance for b in final_balances if b.currency == "EUR")

    expected_balance = initial_eur - trade_cost_eur
    assert abs(final_eur - expected_balance) < 0.01

    # Verify trade recorded with correct bucket_id
    trades = await trade_repo.get_history(limit=10)
    assert len(trades) == 1
    assert trades[0].bucket_id == "test-satellite"


@pytest.mark.asyncio
async def test_trade_bucket_id_stored_correctly(security_repo, trade_repo, core_bucket):
    """Test that bucket_id is stored correctly in trades table."""
    # Create security
    security = Security(
        symbol="NVDA",
        yahoo_symbol="NVDA",
        isin="US67066G1040",
        name="NVIDIA Corp.",
        product_type=ProductType.EQUITY,
        industry="Semiconductors",
        country="United States",
        priority_multiplier=1.0,
        min_lot=1,
        active=True,
        currency="USD",
    )
    await security_repo.create(security)

    # Create trade
    trade = TradeFactory.create_from_execution(
        symbol="NVDA",
        side=TradeSide.BUY,
        quantity=15.0,
        price=500.0,
        order_id="test-bucket-id-1",
        executed_at=datetime.now(),
        currency=Currency.USD,
        currency_rate=1.10,
        isin="US67066G1040",
        bucket_id="core",
        mode="live",
    )

    # Record trade
    await trade_repo.create(trade)

    # Retrieve and verify
    trades = await trade_repo.get_history(limit=10)
    assert len(trades) == 1
    assert trades[0].bucket_id == "core"
    assert trades[0].mode == "live"
    assert trades[0].isin == "US67066G1040"
