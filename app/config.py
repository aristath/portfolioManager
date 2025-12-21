"""Application configuration from environment variables."""

import os
from pathlib import Path
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings loaded from environment."""

    # Application
    app_name: str = "Arduino Trader"
    debug: bool = False

    # Database
    database_path: Path = Path("data/trader.db")

    # Tradernet API
    tradernet_api_key: str = ""
    tradernet_api_secret: str = ""
    tradernet_base_url: str = "https://api.tradernet.com"

    # Scheduling
    daily_sync_hour: int = 18  # Hour for daily portfolio sync
    cash_check_interval_minutes: int = 15  # Check cash balance every 15 min

    # LED Display
    led_serial_port: str = "/dev/ttyACM0"
    led_baud_rate: int = 115200

    # Investment / Rebalancing
    min_cash_threshold: float = 400.0  # EUR - minimum cash to trigger rebalance
    min_trade_size: float = 400.0  # EUR - keeps commission at 0.5% (€2 fee)
    max_trades_per_cycle: int = 5  # Maximum trades per rebalance cycle
    min_stock_score: float = 0.5  # Minimum score to consider buying a stock

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
