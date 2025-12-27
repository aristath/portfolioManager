#!/usr/bin/env python3
"""Native LED Display Script for Arduino Uno Q.

Polls the FastAPI application for display text and sends it to the MCU via Router Bridge.
Runs as a systemd service, replacing the Docker-based Arduino App framework implementation.
"""

import logging
import sys
import time
from pathlib import Path
from typing import Optional

import requests

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from app.config import settings  # noqa: E402

# Configure logging
log_dir = settings.data_dir / "logs"
log_dir.mkdir(parents=True, exist_ok=True)

file_handler = logging.FileHandler(log_dir / "led-display.log")
file_handler.setFormatter(
    logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
)

console_handler = logging.StreamHandler()
console_handler.setFormatter(
    logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")
)

logging.basicConfig(
    level=logging.INFO,
    handlers=[file_handler, console_handler],
)

logger = logging.getLogger(__name__)

# Configuration
API_URL = "http://localhost:8000"
POLL_INTERVAL = 2.0  # Poll every 2 seconds
API_RETRY_DELAY = 5.0  # Retry API connection after 5 seconds
DEFAULT_TICKER_SPEED = 50  # ms per scroll step

# Try to import Router Bridge client
try:
    # Import from scripts directory
    import importlib.util

    router_bridge_path = Path(__file__).parent / "router_bridge_client.py"
    spec = importlib.util.spec_from_file_location("router_bridge_client", router_bridge_path)
    if spec and spec.loader:
        router_bridge_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(router_bridge_module)
        bridge_call = router_bridge_module.call
        ROUTER_BRIDGE_AVAILABLE = True
    else:
        raise ImportError("Failed to load router_bridge_client module")
except ImportError as e:
    logger.warning(f"Router Bridge client not available: {e}")
    ROUTER_BRIDGE_AVAILABLE = False
    bridge_call = None

_session = requests.Session()
_last_text = ""


def set_text(text: str, speed: int = DEFAULT_TICKER_SPEED) -> bool:
    """Send text to MCU for scrolling via Router Bridge.

    Args:
        text: Text to scroll (ASCII only, Euro symbol will be replaced)
        speed: Milliseconds per scroll step (lower = faster)

    Returns:
        True if successful, False otherwise
    """
    if not ROUTER_BRIDGE_AVAILABLE:
        logger.error("Router Bridge client not available")
        return False

    if not text:
        return True  # Empty text is valid

    try:
        # Replace Euro symbol with EUR (Font_5x7 only has ASCII 32-126)
        text = text.replace("€", "EUR")
        bridge_call("scrollText", text, speed, timeout=30)
        return True
    except Exception as e:
        logger.error(f"Failed to set text via Router Bridge: {e}")
        return False


def fetch_display_data() -> Optional[dict]:
    """Fetch display state from API. Returns None on error."""
    try:
        # Fetch full display state which includes ticker_text and ticker_speed
        response = _session.get(f"{API_URL}/api/status/led/display", timeout=10)
        if response.status_code == 200:
            return response.json()
        else:
            logger.warning(f"API returned status {response.status_code}")
            return None
    except requests.exceptions.ConnectionError:
        logger.warning("API connection failed, will retry")
        return None
    except requests.exceptions.Timeout:
        logger.warning("API request timed out")
        return None
    except Exception as e:
        logger.error(f"Unexpected error fetching display data: {e}")
        return None


def _process_display_data(data: dict) -> None:
    """Process display data and update MCU if changed."""
    global _last_text

    # Get ticker text from display state
    ticker_text = data.get("ticker_text", "")
    ticker_speed = data.get("ticker_speed", DEFAULT_TICKER_SPEED)

    # Handle different display modes
    mode = data.get("mode", "normal")
    error_message = data.get("error_message")
    activity_message = data.get("activity_message")

    # Determine what text to display (priority: error > activity > ticker)
    if mode == "error" and error_message:
        display_text = error_message
    elif activity_message:
        display_text = activity_message
    elif ticker_text:
        display_text = ticker_text
    else:
        display_text = ""

    if display_text != _last_text:
        logger.info(f"Updating display text: {display_text[:50]}...")
        if set_text(display_text, speed=ticker_speed):
            _last_text = display_text
        else:
            logger.error("Failed to update display text")


def main_loop():
    """Main loop - fetch display text from API, update MCU via Router Bridge."""
    global _last_text

    logger.info("Starting LED display native script")
    logger.info(f"API URL: {API_URL}")

    if not ROUTER_BRIDGE_AVAILABLE:
        logger.error("Router Bridge client not available. Exiting.")
        sys.exit(1)

    # Test Router Bridge connection
    try:
        bridge_call("scrollText", "TEST", 50, timeout=2)
        logger.info("Router Bridge connection test successful")
    except Exception as e:
        logger.error(f"Router Bridge connection test failed: {e}")
        logger.error("Make sure arduino-router service is running")
        sys.exit(1)

    consecutive_errors = 0
    max_consecutive_errors = 10

    while True:
        try:
            # Fetch data from API
            data = fetch_display_data()

            if data is None:
                # API unavailable - show error on display
                if _last_text != "API OFFLINE":
                    set_text("API OFFLINE", speed=DEFAULT_TICKER_SPEED)
                    _last_text = "API OFFLINE"
                consecutive_errors += 1
                if consecutive_errors >= max_consecutive_errors:
                    logger.error("Too many consecutive API errors, exiting")
                    sys.exit(1)
                time.sleep(API_RETRY_DELAY)
                continue

            consecutive_errors = 0
            _process_display_data(data)

            time.sleep(POLL_INTERVAL)

        except KeyboardInterrupt:
            logger.info("Received interrupt signal, shutting down...")
            break
        except Exception as e:
            logger.error(f"Unexpected error in main loop: {e}", exc_info=True)
            time.sleep(POLL_INTERVAL)


if __name__ == "__main__":
    try:
        main_loop()
    except Exception as e:
        logger.error(f"Fatal error: {e}", exc_info=True)
        sys.exit(1)
