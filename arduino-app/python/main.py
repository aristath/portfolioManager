# Arduino Trader LED Display
# Simple scrolling text display for 8x13 LED matrix and RGB LEDs 3 & 4

from arduino.app_utils import App, Bridge, Logger
import time
import requests

logger = Logger("trader-display")

API_URL = "http://172.17.0.1:8000"

# Persistent HTTP session for connection pooling (reuses TCP connections)
_http_session = requests.Session()

# Default ticker speed in ms (can be overridden by API)
DEFAULT_TICKER_SPEED = 50

# Track last values to avoid unnecessary updates
_last_text = ""
_last_text_speed = 0
_last_led3 = None
_last_led4 = None


def scroll_text(text: str, speed: int = 50) -> bool:
    """Scroll text across LED matrix using native ArduinoGraphics.

    Args:
        text: Text to scroll (uses Euro symbol € directly)
        speed: Milliseconds per scroll step (lower = faster)

    Returns:
        True if successful, False otherwise
    """
    try:
        # Use longer timeout - scrollText may take time to set up scrolling animation
        # Estimate timeout based on text length: ~100ms per character + 5s base
        estimated_timeout = max(10, min(60, len(text) * 0.1 + 5))
        Bridge.call("scrollText", text, speed, timeout=int(estimated_timeout))
        return True
    except TimeoutError:
        logger.warning(f"scrollText timed out: {text[:50]}...")
        return False
    except Exception as e:
        logger.debug(f"scrollText failed: {e}")
        return False


def set_rgb3(r: int, g: int, b: int) -> bool:
    """Set RGB LED 3 color (sync indicator).

    Args:
        r: Red value (0-255, 0 = off, >0 = on)
        g: Green value (0-255)
        b: Blue value (0-255)

    Returns:
        True if successful, False otherwise
    """
    try:
        Bridge.call("setRGB3", r, g, b, timeout=2)
        return True
    except Exception as e:
        logger.debug(f"setRGB3 failed: {e}")
        return False


def set_rgb4(r: int, g: int, b: int) -> bool:
    """Set RGB LED 4 color (processing indicator).

    Args:
        r: Red value (0-255, 0 = off, >0 = on)
        g: Green value (0-255)
        b: Blue value (0-255)

    Returns:
        True if successful, False otherwise
    """
    try:
        Bridge.call("setRGB4", r, g, b, timeout=2)
        return True
    except Exception as e:
        logger.debug(f"setRGB4 failed: {e}")
        return False


def fetch_display_state() -> dict | None:
    """Fetch display state from FastAPI backend.

    Uses persistent session for HTTP connection pooling (keep-alive).
    """
    try:
        with _http_session.get(f"{API_URL}/api/status/led/display", timeout=2) as resp:
            if resp.status_code == 200:
                return resp.json()
    except Exception as e:
        logger.debug(f"API fetch: {e}")
    return None


def estimate_scroll_duration(text: str, speed_ms: int) -> float:
    """Estimate how long scrolling text will take to complete.

    Args:
        text: Text to scroll
        speed_ms: Milliseconds per scroll step

    Returns:
        Estimated duration in seconds
    """
    # Matrix width is 13 pixels, each character is ~5 pixels wide
    # Text needs to scroll from right edge (x=13) to completely off left edge
    # Total scroll distance = text_width + matrix_width
    # Each scroll step moves 1 pixel, takes speed_ms milliseconds
    matrix_width = 13
    char_width = 5
    text_width = len(text) * char_width
    total_steps = text_width + matrix_width
    duration_seconds = (total_steps * speed_ms) / 1000.0
    # Add some buffer for safety
    return duration_seconds + 2.0


def loop():
    """Main loop - fetch display state from API, update MCU if changed.

    Strategy:
    - If no message to display: poll every 10 seconds
    - If we have a message: call scrollText, then wait for estimated completion
      before polling for next message
    """
    global _last_text, _last_text_speed, _last_led3, _last_led4

    try:
        state = fetch_display_state()

        if state is None:
            # API unreachable - show error
            if _last_text != "API OFFLINE":
                if scroll_text("API OFFLINE", DEFAULT_TICKER_SPEED):
                    _last_text = "API OFFLINE"
                    _last_text_speed = DEFAULT_TICKER_SPEED
                    # Wait for scroll to complete before next poll
                    scroll_duration = estimate_scroll_duration("API OFFLINE", DEFAULT_TICKER_SPEED)
                    time.sleep(scroll_duration)
                    return
            # No message to display - poll less frequently
            time.sleep(10)
            return

        # Get state values
        error_message = state.get("error_message")
        activity_message = state.get("activity_message")
        ticker_text = state.get("ticker_text", "")
        ticker_speed = int(state.get("ticker_speed", DEFAULT_TICKER_SPEED))
        led3 = state.get("led3", [0, 0, 0])
        led4 = state.get("led4", [0, 0, 0])

        # Update RGB LEDs (only if changed)
        led3_tuple = tuple(led3)
        if _last_led3 != led3_tuple:
            set_rgb3(led3[0], led3[1], led3[2])
            _last_led3 = led3_tuple

        led4_tuple = tuple(led4)
        if _last_led4 != led4_tuple:
            set_rgb4(led4[0], led4[1], led4[2])
            _last_led4 = led4_tuple

        # Determine what text to display (priority: error > activity > ticker)
        if error_message:
            display_text = error_message
        elif activity_message:
            display_text = activity_message
        elif ticker_text:
            display_text = ticker_text
        else:
            display_text = ""

        # If we have text to display and it's different, call scrollText
        if display_text and (display_text != _last_text or ticker_speed != _last_text_speed):
            if scroll_text(display_text, ticker_speed):
                _last_text = display_text
                _last_text_speed = ticker_speed
                # Wait for scroll animation to complete before polling for next message
                scroll_duration = estimate_scroll_duration(display_text, ticker_speed)
                logger.debug(f"Scrolling '{display_text[:30]}...' for ~{scroll_duration:.1f}s")
                time.sleep(scroll_duration)
                return
            # If scroll_text fails, don't update _last_text so it will retry next iteration
            # But don't wait - retry soon
            time.sleep(2)
            return

        # No message to display or same message - poll less frequently
        time.sleep(10)

    except Exception as e:
        logger.error(f"Loop error: {e}")
        time.sleep(10)


logger.info("LED Display starting (scrolling text mode)...")
App.run(user_loop=loop)
