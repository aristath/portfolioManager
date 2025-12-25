"""Domain constants for business logic."""

# Position sizing multipliers
MIN_CONVICTION_MULTIPLIER = 0.8
MAX_CONVICTION_MULTIPLIER = 1.2
MIN_PRIORITY_MULTIPLIER = 0.9
MAX_PRIORITY_MULTIPLIER = 1.1
MIN_VOLATILITY_MULTIPLIER = 0.7
MAX_POSITION_SIZE_MULTIPLIER = 1.2

# Currency codes - Use Currency enum from app.domain.value_objects.currency

# Trade sides - Use TradeSide enum from app.domain.value_objects.trade_side

# Cooldown periods (days)
BUY_COOLDOWN_DAYS = 30

# Geography codes
GEO_EU = "EU"
GEO_ASIA = "ASIA"
GEO_US = "US"


