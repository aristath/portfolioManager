"""Domain constants for business logic."""

# Position sizing multipliers (legacy conviction-based)
MIN_CONVICTION_MULTIPLIER = 0.8
MAX_CONVICTION_MULTIPLIER = 1.2
MIN_PRIORITY_MULTIPLIER = 0.9
MAX_PRIORITY_MULTIPLIER = 1.1
MIN_VOLATILITY_MULTIPLIER = 0.7
MAX_POSITION_SIZE_MULTIPLIER = 1.2

# Risk Parity Position Sizing
# Based on MOSEK Portfolio Cookbook principles - inverse volatility weighting
TARGET_PORTFOLIO_VOLATILITY = 0.15  # 15% annual target volatility
MIN_VOLATILITY_FOR_SIZING = 0.05    # Floor to prevent division issues
MAX_VOL_WEIGHT = 2.0                # Max 2x base size for low-vol stocks
MIN_VOL_WEIGHT = 0.5                # Min 0.5x base size for high-vol stocks
DEFAULT_VOLATILITY = 0.20           # Default if volatility unknown

# Rebalancing Bands
# Only rebalance when position deviates significantly from target
REBALANCE_BAND_PCT = 0.07  # 7% deviation triggers rebalance

# Currency codes - Use Currency enum from app.domain.value_objects.currency

# Trade sides - Use TradeSide enum from app.domain.value_objects.trade_side

# Cooldown periods (days)
BUY_COOLDOWN_DAYS = 30

# Geography codes
GEO_EU = "EU"
GEO_ASIA = "ASIA"
GEO_US = "US"


