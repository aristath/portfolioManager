"""Unit tests for satellites domain models."""

import pytest

from app.modules.satellites.domain.enums import (
    BucketStatus,
    BucketType,
    TransactionType,
)
from app.modules.satellites.domain.models import (
    Bucket,
    BucketBalance,
    BucketTransaction,
    SatelliteSettings,
)


class TestBucket:
    """Tests for Bucket model."""

    def test_create_core_bucket(self):
        """Test creating a core bucket."""
        bucket = Bucket(
            id="core",
            name="Core Portfolio",
            type=BucketType.CORE,
            status=BucketStatus.ACTIVE,
        )

        assert bucket.id == "core"
        assert bucket.name == "Core Portfolio"
        assert bucket.type == BucketType.CORE
        assert bucket.status == BucketStatus.ACTIVE
        assert bucket.target_pct is None

    def test_create_satellite_bucket(self):
        """Test creating a satellite bucket."""
        bucket = Bucket(
            id="momentum_1",
            name="Momentum Hunter",
            type=BucketType.SATELLITE,
            status=BucketStatus.RESEARCH,
            notes="Testing momentum strategy",
            target_pct=0.05,
            min_pct=0.02,
            max_pct=0.15,
        )

        assert bucket.id == "momentum_1"
        assert bucket.type == BucketType.SATELLITE
        assert bucket.status == BucketStatus.RESEARCH
        assert bucket.target_pct == 0.05
        assert bucket.min_pct == 0.02
        assert bucket.max_pct == 0.15

    def test_bucket_default_values(self):
        """Test bucket default values."""
        bucket = Bucket(
            id="test",
            name="Test",
            type=BucketType.SATELLITE,
            status=BucketStatus.RESEARCH,
        )

        assert bucket.consecutive_losses == 0
        assert bucket.max_consecutive_losses == 5
        assert bucket.high_water_mark == 0.0

    def test_bucket_status_values(self):
        """Test all bucket status values."""
        statuses = [
            BucketStatus.RESEARCH,
            BucketStatus.ACCUMULATING,
            BucketStatus.ACTIVE,
            BucketStatus.HIBERNATING,
            BucketStatus.PAUSED,
            BucketStatus.RETIRED,
        ]

        for status in statuses:
            bucket = Bucket(
                id="test",
                name="Test",
                type=BucketType.SATELLITE,
                status=status,
            )
            assert bucket.status == status


class TestBucketBalance:
    """Tests for BucketBalance model."""

    def test_create_balance(self):
        """Test creating a bucket balance."""
        balance = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=1000.50,
            last_updated="2024-01-15T10:30:00",
        )

        assert balance.bucket_id == "core"
        assert balance.currency == "EUR"
        assert balance.balance == 1000.50

    def test_balance_multiple_currencies(self):
        """Test balances for different currencies."""
        eur_balance = BucketBalance(
            bucket_id="core",
            currency="EUR",
            balance=1000.0,
            last_updated="2024-01-15",
        )
        usd_balance = BucketBalance(
            bucket_id="core",
            currency="USD",
            balance=500.0,
            last_updated="2024-01-15",
        )

        assert eur_balance.currency == "EUR"
        assert usd_balance.currency == "USD"


class TestBucketTransaction:
    """Tests for BucketTransaction model."""

    def test_create_deposit_transaction(self):
        """Test creating a deposit transaction."""
        tx = BucketTransaction(
            bucket_id="core",
            type=TransactionType.DEPOSIT,
            amount=1000.0,
            currency="EUR",
            description="Monthly deposit",
        )

        assert tx.bucket_id == "core"
        assert tx.type == TransactionType.DEPOSIT
        assert tx.amount == 1000.0
        assert tx.currency == "EUR"

    def test_create_trade_transaction(self):
        """Test creating trade transactions."""
        buy_tx = BucketTransaction(
            bucket_id="satellite_1",
            type=TransactionType.TRADE_BUY,
            amount=-500.0,
            currency="EUR",
            description="Buy AAPL",
        )
        sell_tx = BucketTransaction(
            bucket_id="satellite_1",
            type=TransactionType.TRADE_SELL,
            amount=550.0,
            currency="EUR",
            description="Sell AAPL",
        )

        assert buy_tx.type == TransactionType.TRADE_BUY
        assert buy_tx.amount < 0  # Buy reduces cash
        assert sell_tx.type == TransactionType.TRADE_SELL
        assert sell_tx.amount > 0  # Sell increases cash

    def test_transaction_types(self):
        """Test all transaction types."""
        types = [
            TransactionType.DEPOSIT,
            TransactionType.REALLOCATION,
            TransactionType.TRADE_BUY,
            TransactionType.TRADE_SELL,
            TransactionType.DIVIDEND,
            TransactionType.TRANSFER_IN,
            TransactionType.TRANSFER_OUT,
        ]

        for tx_type in types:
            tx = BucketTransaction(
                bucket_id="test",
                type=tx_type,
                amount=100.0,
                currency="EUR",
            )
            assert tx.type == tx_type


class TestSatelliteSettings:
    """Tests for SatelliteSettings model."""

    def test_create_default_settings(self):
        """Test creating settings with defaults."""
        settings = SatelliteSettings(satellite_id="momentum_1")

        assert settings.satellite_id == "momentum_1"
        assert settings.risk_appetite == 0.5
        assert settings.hold_duration == 0.5
        assert settings.entry_style == 0.5
        assert settings.position_spread == 0.5
        assert settings.profit_taking == 0.5
        assert settings.trailing_stops is False
        assert settings.follow_regime is False
        assert settings.auto_harvest is False
        assert settings.pause_high_volatility is False
        assert settings.dividend_handling == "reinvest_same"

    def test_create_custom_settings(self):
        """Test creating settings with custom values."""
        settings = SatelliteSettings(
            satellite_id="aggressive_1",
            preset="momentum_hunter",
            risk_appetite=0.8,
            hold_duration=0.3,
            entry_style=0.7,
            position_spread=0.4,
            profit_taking=0.6,
            trailing_stops=True,
            follow_regime=True,
            auto_harvest=False,
            pause_high_volatility=True,
            dividend_handling="send_to_core",
        )

        assert settings.risk_appetite == 0.8
        assert settings.hold_duration == 0.3
        assert settings.trailing_stops is True
        assert settings.dividend_handling == "send_to_core"

    def test_validate_valid_settings(self):
        """Test validation passes for valid settings."""
        settings = SatelliteSettings(
            satellite_id="test",
            risk_appetite=0.0,
            hold_duration=1.0,
        )
        settings.validate()  # Should not raise

    def test_validate_slider_out_of_range(self):
        """Test validation fails for out-of-range sliders."""
        settings = SatelliteSettings(
            satellite_id="test",
            risk_appetite=1.5,  # Out of range
        )
        with pytest.raises(ValueError, match="must be between 0.0 and 1.0"):
            settings.validate()

    def test_validate_negative_slider(self):
        """Test validation fails for negative slider values."""
        settings = SatelliteSettings(
            satellite_id="test",
            hold_duration=-0.1,  # Negative
        )
        with pytest.raises(ValueError, match="must be between 0.0 and 1.0"):
            settings.validate()

    def test_validate_invalid_dividend_handling(self):
        """Test validation fails for invalid dividend handling."""
        settings = SatelliteSettings(
            satellite_id="test",
            dividend_handling="invalid_option",
        )
        with pytest.raises(ValueError, match="dividend_handling must be one of"):
            settings.validate()

    def test_valid_dividend_handling_options(self):
        """Test all valid dividend handling options."""
        valid_options = ["reinvest_same", "send_to_core", "accumulate_cash"]

        for option in valid_options:
            settings = SatelliteSettings(
                satellite_id="test",
                dividend_handling=option,
            )
            settings.validate()  # Should not raise
