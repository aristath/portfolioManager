"""Settings API endpoints."""

from typing import Any, Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.core.cache.cache import cache
from app.infrastructure.dependencies import (
    CalculationsRepositoryDep,
    DatabaseManagerDep,
    SettingsRepositoryDep,
)

router = APIRouter()


# Default values for all configurable settings
# NOTE: Planner settings have been moved to TOML configuration files.
# Transaction costs, trading constraints, and all planner algorithms are now
# per-bucket and configured via config/planner/*.toml files.
SETTING_DEFAULTS = {
    # Security scoring
    "min_security_score": 0.5,  # Minimum score for security to be recommended (0-1)
    "target_annual_return": 0.11,  # Optimal CAGR for scoring (11%)
    "market_avg_pe": 22.0,  # Reference P/E for valuation
    # Trading mode
    "trading_mode": "research",  # "live" or "research" - blocks trades in research mode
    # Portfolio Optimizer settings
    "optimizer_blend": 0.5,  # 0.0 = pure Mean-Variance, 1.0 = pure HRP
    "optimizer_target_return": 0.11,  # Target annual return for MV component
    # Cash management
    "min_cash_reserve": 500.0,  # Minimum cash to keep (never fully deploy)
    # LED Matrix settings
    "ticker_speed": 50.0,  # Ticker scroll speed in ms per frame (lower = faster)
    "led_brightness": 150.0,  # LED brightness (0-255)
    # Ticker display options (1.0 = show, 0.0 = hide)
    "ticker_show_value": 1.0,  # Show portfolio value
    "ticker_show_cash": 1.0,  # Show cash balance
    "ticker_show_actions": 1.0,  # Show next actions (BUY/SELL)
    "ticker_show_amounts": 1.0,  # Show amounts for actions
    "ticker_max_actions": 3.0,  # Max recommendations to show (buy + sell)
    # Job scheduling intervals (simplified to 3 configurable settings)
    "job_sync_cycle_minutes": 15.0,  # Unified sync cycle interval (trades, prices, recommendations)
    "job_maintenance_hour": 3.0,  # Daily maintenance hour (0-23)
    "job_auto_deploy_minutes": 5.0,  # Auto-deploy check interval (minutes)
    # Universe Pruning settings
    "universe_pruning_enabled": 1.0,  # 1.0 = enabled, 0.0 = disabled
    "universe_pruning_score_threshold": 0.50,  # Minimum average score to keep security (0-1)
    "universe_pruning_months": 3.0,  # Number of months to look back for scores
    "universe_pruning_min_samples": 2.0,  # Minimum number of score samples required
    "universe_pruning_check_delisted": 1.0,  # 1.0 = check for delisted securities, 0.0 = skip
    # Event-Driven Rebalancing settings
    "event_driven_rebalancing_enabled": 1.0,  # 1.0 = enabled, 0.0 = disabled
    "rebalance_position_drift_threshold": 0.05,  # Position drift threshold (0.05 = 5%)
    "rebalance_cash_threshold_multiplier": 2.0,  # Cash threshold = multiplier × min_trade_size
    # Trade Frequency Limits settings
    "trade_frequency_limits_enabled": 1.0,  # 1.0 = enabled, 0.0 = disabled
    "min_time_between_trades_minutes": 60.0,  # Minimum minutes between any trades
    "max_trades_per_day": 4.0,  # Maximum trades per calendar day
    "max_trades_per_week": 10.0,  # Maximum trades per rolling 7-day window
    # Security Discovery settings
    "security_discovery_enabled": 1.0,  # 1.0 = enabled, 0.0 = disabled
    "security_discovery_score_threshold": 0.75,  # Minimum score to add security (0-1)
    "security_discovery_max_per_month": 2.0,  # Maximum securities to add per month
    "security_discovery_require_manual_review": 0.0,  # 1.0 = require review, 0.0 = auto-add
    "security_discovery_geographies": "EU,US,ASIA",  # Comma-separated geography list
    "security_discovery_exchanges": "usa,europe",  # Comma-separated exchange list
    "security_discovery_min_volume": 1000000.0,  # Minimum daily volume for liquidity
    "security_discovery_fetch_limit": 50.0,  # Maximum candidates to fetch from API
    # Market Regime Detection settings
    "market_regime_detection_enabled": 1.0,  # 1.0 = enabled, 0.0 = disabled
    "market_regime_bull_cash_reserve": 0.02,  # Cash reserve percentage in bull market (2%)
    "market_regime_bear_cash_reserve": 0.05,  # Cash reserve percentage in bear market (5%)
    "market_regime_sideways_cash_reserve": 0.03,  # Cash reserve percentage in sideways market (3%)
    "market_regime_bull_threshold": 0.05,  # Threshold for bull market (5% above MA)
    "market_regime_bear_threshold": -0.05,  # Threshold for bear market (-5% below MA)
}


class SettingUpdate(BaseModel):
    value: float


async def get_setting(
    key: str, settings_repo: SettingsRepositoryDep, default: Optional[str] = None
) -> str | None:
    """Get a setting value from the database."""
    value = await settings_repo.get(key)
    return str(value) if value is not None else default


async def get_settings_batch(
    keys: list[str], settings_repo: SettingsRepositoryDep
) -> dict[str, str]:
    """Get multiple settings in a single database query (cached 3s)."""
    cache_key = "settings:all"
    cached = cache.get(cache_key)
    if cached is not None:
        # Return only requested keys from cached data
        return {k: v for k, v in cached.items() if k in keys}

    # Fetch all settings from DB
    all_settings = await settings_repo.get_all()

    # Cache for 3 seconds
    cache.set(cache_key, all_settings, ttl_seconds=3)

    return {k: v for k, v in all_settings.items() if k in keys}


async def set_setting(
    key: str, value: str, settings_repo: SettingsRepositoryDep
) -> None:
    """Set a setting value in the database."""
    await settings_repo.set_float(key, float(value))
    # Invalidate settings cache
    cache.invalidate("settings:all")


async def get_setting_value(key: str, settings_repo: SettingsRepositoryDep) -> float:
    """Get a setting value, falling back to default."""
    db_value = await get_setting(key, settings_repo)
    if db_value:
        return float(db_value)
    default_val = SETTING_DEFAULTS.get(key, 0)
    return float(default_val) if isinstance(default_val, (int, float)) else 0.0


async def get_trading_mode(settings_repo: SettingsRepositoryDep) -> str:
    """Get trading mode setting (returns "live" or "research")."""
    db_value = await get_setting("trading_mode", settings_repo)
    if db_value in ("live", "research"):
        return db_value
    return str(SETTING_DEFAULTS.get("trading_mode", "research"))


async def set_trading_mode(mode: str, settings_repo: SettingsRepositoryDep) -> None:
    """Set trading mode setting (must be "live" or "research")."""
    if mode not in ("live", "research"):
        raise ValueError(f"Invalid trading mode: {mode}. Must be 'live' or 'research'")
    await settings_repo.set("trading_mode", mode, "Trading mode: 'live' or 'research'")
    # Invalidate settings cache
    cache.invalidate("settings:all")


async def get_job_settings(settings_repo: SettingsRepositoryDep) -> dict[str, float]:
    """Get all job scheduling settings in one query."""
    job_keys = [k for k in SETTING_DEFAULTS if k.startswith("job_")]
    db_values = await get_settings_batch(job_keys, settings_repo)
    result = {}
    for key in job_keys:
        if key in db_values:
            val = db_values[key]
            result[key] = float(val) if isinstance(val, (int, float, str)) else 0.0
        else:
            default_val = SETTING_DEFAULTS[key]
            result[key] = (
                float(default_val) if isinstance(default_val, (int, float)) else 0.0
            )
    return result


@router.get("")
async def get_all_settings(settings_repo: SettingsRepositoryDep):
    """Get all configurable settings."""
    # Get all settings in a single query
    keys = list(SETTING_DEFAULTS.keys())
    db_values = await get_settings_batch(keys, settings_repo)

    # String settings that should be returned as strings
    string_settings = {
        "trading_mode",
        "security_discovery_geographies",
        "security_discovery_exchanges",
        "risk_profile",
    }

    result: dict[str, Any] = {}
    for key, default in SETTING_DEFAULTS.items():
        if key in db_values:
            if key in string_settings:
                result[key] = str(db_values[key])
            else:
                val = db_values[key]
                result[key] = float(val) if isinstance(val, (int, float, str)) else 0.0
        else:
            if key in string_settings:
                result[key] = str(default) if default is not None else ""
            else:
                default_val = default if isinstance(default, (int, float)) else 0.0
                result[key] = float(default_val)
    return result


@router.put("/{key}")
async def update_setting_value(
    key: str, data: SettingUpdate, settings_repo: SettingsRepositoryDep
):
    """Update a setting value."""
    if key not in SETTING_DEFAULTS:
        raise HTTPException(status_code=400, detail=f"Unknown setting: {key}")

    # Special handling for string settings
    if key == "trading_mode":
        mode = str(data.value).lower()
        if mode not in ("live", "research"):
            raise HTTPException(
                status_code=400,
                detail=f"Invalid trading mode: {mode}. Must be 'live' or 'research'",
            )
        await set_trading_mode(mode, settings_repo)
        return {key: mode}
    elif key in ("security_discovery_geographies", "security_discovery_exchanges"):
        # Store as string for comma-separated lists
        await settings_repo.set(key, str(data.value))
        cache.invalidate("settings:all")
        return {key: str(data.value)}
    elif key in (
        "market_regime_bull_cash_reserve",
        "market_regime_bear_cash_reserve",
        "market_regime_sideways_cash_reserve",
    ):
        # Validate percentage range (1% to 40%)
        if data.value < 0.01 or data.value > 0.40:
            raise HTTPException(
                status_code=400,
                detail=f"{key} must be between 1% (0.01) and 40% (0.40)",
            )
        await set_setting(key, str(data.value), settings_repo)
        return {key: data.value}

    # All planner settings have been moved to TOML configuration
    # Generic validation for remaining settings
    await set_setting(key, str(data.value), settings_repo)

    # Invalidate recommendation caches when recommendation-affecting settings change
    # Note: Planner settings now come from TOML and don't need cache invalidation
    # Only settings that affect security scoring or portfolio optimization remain here
    recommendation_settings = {
        "min_security_score",  # Security scoring
        "target_annual_return",  # Security scoring
        "optimizer_blend",  # Portfolio optimizer
        "optimizer_target_return",  # Portfolio optimizer
        "min_cash_reserve",  # Global trading constraint
    }
    if key in recommendation_settings:
        from app.infrastructure.recommendation_cache import get_recommendation_cache

        # Invalidate SQLite recommendation cache
        rec_cache = get_recommendation_cache()
        await rec_cache.invalidate_all_recommendations()

        # Invalidate in-memory caches
        cache.invalidate_prefix("recommendations")  # Unified recommendations cache

    return {key: data.value}


@router.post("/restart-service")
async def restart_service():
    """Restart the arduino-trader systemd service."""
    import subprocess

    try:
        result = subprocess.run(
            ["sudo", "systemctl", "restart", "arduino-trader"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode == 0:
            return {"status": "ok", "message": "Service restart initiated"}
        else:
            return {
                "status": "error",
                "message": f"Failed to restart service: {result.stderr}",
            }
    except Exception as e:
        return {"status": "error", "message": str(e)}


@router.post("/restart")
async def restart_system():
    """Trigger system reboot."""
    import subprocess

    subprocess.Popen(["sudo", "reboot"])
    return {"status": "rebooting"}


@router.post("/reset-cache")
async def reset_cache():
    """Clear all cached data including score cache."""
    from app.infrastructure.recommendation_cache import get_recommendation_cache

    # Clear simple in-memory cache
    cache.clear()

    # Clear SQLite recommendation cache
    rec_cache = get_recommendation_cache()
    await rec_cache.invalidate_all_recommendations()

    # Metrics expire naturally via TTL, no manual invalidation needed

    return {"status": "ok", "message": "All caches cleared"}


@router.get("/cache-stats")
async def get_cache_stats(
    calc_repo: CalculationsRepositoryDep,
    db_manager: DatabaseManagerDep,
):
    """Get cache statistics."""
    db = db_manager.calculations

    # Get calculations.db stats
    row = await db.fetchone("SELECT COUNT(*) as cnt FROM calculated_metrics")
    calc_count = row["cnt"] if row else 0

    expired_count = await calc_repo.delete_expired()

    return {
        "simple_cache": {
            "entries": len(cache._cache) if hasattr(cache, "_cache") else 0,
        },
        "calculations_db": {
            "entries": calc_count,
            "expired_cleaned": expired_count,
        },
    }


@router.post("/reschedule-jobs")
async def reschedule_jobs():
    """Reschedule all jobs with current settings values."""
    from app.jobs.scheduler import reschedule_all_jobs

    await reschedule_all_jobs()
    return {"status": "ok", "message": "Jobs rescheduled"}


@router.get("/trading-mode")
async def get_trading_mode_endpoint(settings_repo: SettingsRepositoryDep):
    """Get current trading mode."""
    mode = await get_trading_mode(settings_repo)
    return {"trading_mode": mode}


@router.post("/trading-mode")
async def toggle_trading_mode(settings_repo: SettingsRepositoryDep):
    """Toggle trading mode between 'live' and 'research'."""
    current_mode = await get_trading_mode(settings_repo)
    new_mode = "research" if current_mode == "live" else "live"
    await set_trading_mode(new_mode, settings_repo)
    return {"trading_mode": new_mode, "previous_mode": current_mode}
