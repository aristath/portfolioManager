"""Satellites domain models."""

from dataclasses import dataclass
from datetime import datetime
from typing import Optional

from app.modules.satellites.domain.enums import (
    BucketStatus,
    BucketType,
    TransactionType,
)


@dataclass
class Bucket:
    """A portfolio bucket (core or satellite).

    Attributes:
        id: Unique identifier (e.g., 'core', 'satellite_momentum_1')
        name: Human-readable name
        type: Either 'core' or 'satellite'
        status: Current lifecycle status
        notes: User documentation for the satellite
        target_pct: Target allocation percentage (0.0-1.0)
        min_pct: Minimum percentage before hibernation
        max_pct: Maximum allowed percentage
        consecutive_losses: Count of consecutive losing trades
        max_consecutive_losses: Circuit breaker threshold
        high_water_mark: Peak bucket value for drawdown calculation
        high_water_mark_date: When the high water mark was set
        loss_streak_paused_at: When circuit breaker triggered
        created_at: Creation timestamp
        updated_at: Last update timestamp
    """

    id: str
    name: str
    type: BucketType
    status: BucketStatus = BucketStatus.ACTIVE
    notes: Optional[str] = None
    target_pct: Optional[float] = None
    min_pct: Optional[float] = None
    max_pct: Optional[float] = None
    consecutive_losses: int = 0
    max_consecutive_losses: int = 5
    high_water_mark: float = 0.0
    high_water_mark_date: Optional[str] = None
    loss_streak_paused_at: Optional[str] = None
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    def __post_init__(self) -> None:
        """Convert string values to enums if necessary."""
        if isinstance(self.type, str):
            self.type = BucketType(self.type)
        if isinstance(self.status, str):
            self.status = BucketStatus(self.status)

    @property
    def is_active(self) -> bool:
        """Check if bucket is in an active trading state."""
        return self.status == BucketStatus.ACTIVE

    @property
    def is_trading_allowed(self) -> bool:
        """Check if new trades are allowed."""
        return self.status in (BucketStatus.ACTIVE, BucketStatus.ACCUMULATING)

    @property
    def is_core(self) -> bool:
        """Check if this is the core bucket."""
        return self.type == BucketType.CORE

    @property
    def is_satellite(self) -> bool:
        """Check if this is a satellite bucket."""
        return self.type == BucketType.SATELLITE

    def calculate_drawdown(self, current_value: float) -> float:
        """Calculate current drawdown from high water mark.

        Returns:
            Drawdown as a percentage (0.0-1.0)
        """
        if self.high_water_mark <= 0:
            return 0.0
        return max(0.0, (self.high_water_mark - current_value) / self.high_water_mark)


@dataclass
class BucketBalance:
    """Virtual cash balance for a bucket in a specific currency.

    Attributes:
        bucket_id: ID of the bucket
        currency: Currency code (e.g., 'EUR', 'USD')
        balance: Current balance amount
        last_updated: Timestamp of last update
    """

    bucket_id: str
    currency: str
    balance: float
    last_updated: str


@dataclass
class BucketTransaction:
    """Audit trail entry for bucket cash flow.

    Attributes:
        bucket_id: ID of the bucket
        type: Type of transaction
        amount: Transaction amount (positive for inflow, negative for outflow)
        currency: Currency code
        description: Optional description
        created_at: Transaction timestamp
        id: Database ID (set after insert)
    """

    bucket_id: str
    type: TransactionType
    amount: float
    currency: str
    description: Optional[str] = None
    created_at: Optional[str] = None
    id: Optional[int] = None

    def __post_init__(self) -> None:
        """Convert string values to enums if necessary."""
        if isinstance(self.type, str):
            self.type = TransactionType(self.type)
        if self.created_at is None:
            self.created_at = datetime.now().isoformat()


@dataclass
class SatelliteSettings:
    """Strategy configuration settings for a satellite.

    Slider values are in range 0.0-1.0 where:
    - 0.0 = Conservative/Left option
    - 1.0 = Aggressive/Right option

    Attributes:
        satellite_id: ID of the satellite
        preset: Strategy preset name (e.g., 'momentum_hunter')
        risk_appetite: 0=Conservative, 1=Aggressive
        hold_duration: 0=Quick flips, 1=Patient holds
        entry_style: 0=Buy dips, 1=Buy breakouts
        position_spread: 0=Concentrated, 1=Diversified
        profit_taking: 0=Let winners run, 1=Take profits early
        trailing_stops: Enable trailing stops
        follow_regime: Follow market regime signals
        auto_harvest: Auto-harvest gains to core
        pause_high_volatility: Pause during high volatility
        dividend_handling: How to handle dividends
    """

    satellite_id: str
    preset: Optional[str] = None
    risk_appetite: float = 0.5
    hold_duration: float = 0.5
    entry_style: float = 0.5
    position_spread: float = 0.5
    profit_taking: float = 0.5
    trailing_stops: bool = False
    follow_regime: bool = False
    auto_harvest: bool = False
    pause_high_volatility: bool = False
    dividend_handling: str = "reinvest_same"

    def validate(self) -> None:
        """Validate slider values are in range."""
        for attr in [
            "risk_appetite",
            "hold_duration",
            "entry_style",
            "position_spread",
            "profit_taking",
        ]:
            value = getattr(self, attr)
            if not 0.0 <= value <= 1.0:
                raise ValueError(f"{attr} must be between 0.0 and 1.0, got {value}")

        valid_dividend_options = ["reinvest_same", "send_to_core", "accumulate_cash"]
        if self.dividend_handling not in valid_dividend_options:
            raise ValueError(
                f"dividend_handling must be one of {valid_dividend_options}"
            )
