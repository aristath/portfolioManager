"""
Portfolio Hash Generation

Generates a deterministic hash from current portfolio state.
Used to identify when recommendations apply to the same portfolio state.

The hash includes:
- All positions (including zero quantities for stocks in universe)
- Cash balances as pseudo-positions (CASH.EUR, CASH.USD, etc.)
- The full stocks universe to detect when new stocks are added
- Per-symbol configuration: allow_buy, allow_sell, min_portfolio_target, max_portfolio_target, country, industry
"""

import hashlib
from typing import Any, Dict, List, Optional

from app.domain.models import Stock


def generate_portfolio_hash(
    positions: List[Dict[str, Any]],
    stocks: Optional[List[Stock]] = None,
    cash_balances: Optional[Dict[str, float]] = None,
) -> str:
    """
    Generate a deterministic hash from current portfolio state.

    Args:
        positions: List of position dicts with 'symbol' and 'quantity' keys
        stocks: Optional list of Stock objects in universe (to detect new stocks and include config)
        cash_balances: Optional dict of currency -> amount (e.g., {"EUR": 1500.0})

    Returns:
        8-character hex hash (first 8 chars of MD5)

    Example:
        positions = [{"symbol": "AAPL", "quantity": 10}]
        stocks = [Stock(symbol="AAPL", ...), Stock(symbol="MSFT", ...)]
        cash = {"EUR": 1500.0, "USD": 200.0}
        hash = generate_portfolio_hash(positions, stocks, cash)
    """
    # Build a dict of symbol -> quantity from positions
    # Use float | int to handle both stock quantities (int) and cash (float)
    position_map: Dict[str, float | int] = {
        p["symbol"].upper(): int(p.get("quantity", 0) or 0) for p in positions
    }

    # Build a dict of symbol -> stock config data
    stock_config_map: Dict[str, Dict[str, Any]] = {}

    if stocks:
        for stock in stocks:
            symbol_upper = stock.symbol.upper()
            # Ensure stock is in position_map (with 0 if not held)
            if symbol_upper not in position_map:
                position_map[symbol_upper] = 0

            # Extract config fields
            stock_config_map[symbol_upper] = {
                "allow_buy": stock.allow_buy,
                "allow_sell": stock.allow_sell,
                "min_portfolio_target": stock.min_portfolio_target,
                "max_portfolio_target": stock.max_portfolio_target,
                "country": stock.country or "",
                "industry": stock.industry or "",
            }

    # Add cash balances as pseudo-positions (filter out zero balances)
    if cash_balances:
        for currency, amount in cash_balances.items():
            if amount > 0:
                # Round to 2 decimal places for stability
                position_map[f"CASH.{currency.upper()}"] = round(amount, 2)

    # Sort by symbol for deterministic ordering
    sorted_symbols = sorted(position_map.keys())

    # Build canonical string: "SYMBOL:QUANTITY:allow_buy:allow_sell:min_target:max_target:country:industry"
    parts = []
    for symbol in sorted_symbols:
        quantity = position_map[symbol]
        # Use int for stock quantities, float for cash
        if symbol.startswith("CASH."):
            parts.append(f"{symbol}:{quantity}")
        else:
            # Get config for this symbol (use defaults if not in stocks list)
            config = stock_config_map.get(symbol, {})
            allow_buy = config.get("allow_buy", True)
            allow_sell = config.get("allow_sell", False)
            min_target = config.get("min_portfolio_target")
            max_target = config.get("max_portfolio_target")
            country = config.get("country", "")
            industry = config.get("industry", "")

            # Format config fields
            min_target_str = "" if min_target is None else str(min_target)
            max_target_str = "" if max_target is None else str(max_target)

            parts.append(
                f"{symbol}:{int(quantity)}:{allow_buy}:{allow_sell}:{min_target_str}:{max_target_str}:{country}:{industry}"
            )

    canonical = ",".join(parts)

    # Generate hash and return first 8 characters
    full_hash = hashlib.md5(canonical.encode()).hexdigest()
    return full_hash[:8]


def generate_settings_hash(settings_dict: Dict[str, Any]) -> str:
    """
    Generate a deterministic hash from settings that affect recommendations.

    Args:
        settings_dict: Dictionary of settings values

    Returns:
        8-character hex hash (first 8 chars of MD5)

    Example:
        settings = {"min_trade_size": 100, "min_stock_score": 0.5}
        hash = generate_settings_hash(settings)  # e.g., "b1c2d3e4"
    """
    # Settings that affect recommendation calculations
    # Note: min_trade_size and recommendation_depth removed (handled by optimizer now)
    relevant_keys = sorted(
        [
            "min_stock_score",
            "min_hold_days",
            "sell_cooldown_days",
            "max_loss_threshold",
            "target_annual_return",
            "optimizer_blend",
            "optimizer_target_return",
            "transaction_cost_fixed",
            "transaction_cost_percent",
            "min_cash_reserve",
            "max_plan_depth",
        ]
    )

    # Build canonical string: "key:value,key:value,..."
    parts = [f"{k}:{settings_dict.get(k, '')}" for k in relevant_keys]
    canonical = ",".join(parts)

    # Generate hash and return first 8 characters
    full_hash = hashlib.md5(canonical.encode()).hexdigest()
    return full_hash[:8]


def generate_recommendation_cache_key(
    positions: List[Dict[str, Any]],
    settings_dict: Dict[str, Any],
    stocks: Optional[List[Stock]] = None,
    cash_balances: Optional[Dict[str, float]] = None,
) -> str:
    """
    Generate a cache key from portfolio state and settings.

    This ensures that cache is invalidated when positions, settings,
    stocks universe, cash balances, or per-symbol configuration changes.

    Args:
        positions: List of position dicts with 'symbol' and 'quantity' keys
        settings_dict: Dictionary of settings values
        stocks: Optional list of Stock objects in universe
        cash_balances: Optional dict of currency -> amount

    Returns:
        17-character combined hash (portfolio_hash:settings_hash)

    Example:
        positions = [{"symbol": "AAPL", "quantity": 10}]
        settings = {"min_stock_score": 0.5}
        stocks = [Stock(symbol="AAPL", ...), Stock(symbol="MSFT", ...)]
        cash = {"EUR": 1500.0}
        key = generate_recommendation_cache_key(positions, settings, stocks, cash)
    """
    portfolio_hash = generate_portfolio_hash(positions, stocks, cash_balances)
    settings_hash = generate_settings_hash(settings_dict)
    return f"{portfolio_hash}:{settings_hash}"
