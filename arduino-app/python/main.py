# Arduino Trader LED Display
# Shows portfolio value as big digits + progress bar

from arduino.app_utils import App, Bridge, Logger, Frame
import time
import requests
import numpy as np

logger = Logger("trader-display")

API_URL = "http://172.17.0.1:8000"

ROWS = 8
COLS = 13

# Brightness constants (0 = off, 255 = brightest)
PIXEL_ON = 255    # Lit pixel (brightest)
PIXEL_OFF = 0     # Unlit pixel (off)
PIXEL_DIM = 100   # Dim pixel for heartbeat

# 3x5 digit patterns
DIGITS = {
    '0': ['111','101','101','101','111'],
    '1': ['010','110','010','010','111'],
    '2': ['111','001','111','100','111'],
    '3': ['111','001','111','001','111'],
    '4': ['101','101','111','001','001'],
    '5': ['111','100','111','001','111'],
    '6': ['111','100','111','101','111'],
    '7': ['111','001','001','001','001'],
    '8': ['111','101','111','101','111'],
    '9': ['111','101','111','001','111'],
}

LETTERS = {
    'N': ['101','111','111','101','101'],
    'O': ['111','101','101','101','111'],
    'W': ['101','101','111','111','101'],
    'I': ['111','010','010','010','111'],
    'F': ['111','100','110','100','100'],
}

# State
last_value = 0
scroll_offset = 0
syncing_frame = 0
current_balance_arr = None  # Store current balance display

def create_balance_arr(value):
    """Create big digits + progress bar array for portfolio value.

    Layout:
    - Rows 0-4: Digits (5 rows) + progress bar (5 pixels, each = €200)
    - Row 5: Separator (empty)
    - Rows 6-7: Reserved for status indicators
    """
    arr = np.zeros((ROWS, COLS), dtype=np.uint8)

    # Get thousands and remainder
    thousands = int(value // 1000)
    remainder = int(value % 1000)

    # Convert thousands to string
    value_str = str(thousands) if thousands > 0 else "0"

    # Draw digits (left-aligned, top-aligned at row 0)
    col = 0
    start_row = 0  # Top of display

    for char in value_str:
        if char in DIGITS and col + 3 <= 11:  # Leave room for progress bar
            for row_idx, row_pattern in enumerate(DIGITS[char]):
                for col_idx, pixel in enumerate(row_pattern):
                    if pixel == '1':
                        r = start_row + row_idx
                        c = col + col_idx
                        if 0 <= r < 5 and 0 <= c < COLS:  # Only rows 0-4
                            arr[r][c] = PIXEL_ON
            col += 4  # 3 for digit + 1 space

    # Draw progress bar (column 12 only, rows 0-4, bottom to top)
    # Each pixel = 200 EUR, 5 pixels = 1000 EUR
    full_pixels = remainder // 200
    partial = remainder % 200
    partial_brightness = int((partial / 200) * 255)

    for i in range(5):
        row = 4 - i  # Bottom to top within rows 0-4
        if i < full_pixels:
            arr[row][12] = PIXEL_ON
        elif i == full_pixels and partial_brightness > 0:
            arr[row][12] = partial_brightness

    return arr

def create_balance_frame(value):
    """Create Frame from balance array."""
    return Frame(create_balance_arr(value))

def create_status_frame(brightness=PIXEL_DIM):
    """Create status indicator on bottom 2 rows only.

    Used for heartbeat and other status updates without
    disrupting the main display on rows 0-4.
    """
    arr = np.zeros((ROWS, COLS), dtype=np.uint8)
    # Only light up rows 6-7 (bottom 2 rows)
    arr[6, :] = brightness
    arr[7, :] = brightness
    return Frame(arr)

def clear_status_frame():
    """Clear the status rows (6-7) only."""
    arr = np.zeros((ROWS, COLS), dtype=np.uint8)
    return Frame(arr)

def create_syncing_frame(phase):
    """Create wave animation for syncing."""
    arr = np.zeros((ROWS, COLS), dtype=np.uint8)
    wave_col = phase % (COLS + 4) - 2
    for r in range(ROWS):
        for c in range(COLS):
            dist = abs(c - wave_col)
            if dist <= 2:
                arr[r][c] = max(0, 255 - dist * 80)
    return Frame(arr)

def create_no_wifi_frame(offset):
    """Create scrolling NO WIFI text."""
    arr = np.zeros((ROWS, COLS), dtype=np.uint8)
    text = "NO WIFI"
    text_width = len(text) * 4
    start_col = COLS - (offset % (text_width + COLS))
    start_row = 1

    col = start_col
    for char in text:
        if char == ' ':
            col += 2
            continue
        pattern = LETTERS.get(char, DIGITS.get(char))
        if pattern:
            for row_idx, row_pattern in enumerate(pattern):
                for col_idx, pixel in enumerate(row_pattern):
                    if pixel == '1':
                        r = start_row + row_idx
                        c = col + col_idx
                        if 0 <= r < ROWS and 0 <= c < COLS:
                            arr[r][c] = PIXEL_ON
            col += 4
    return Frame(arr)

def create_error_frame():
    """Create X pattern for errors."""
    arr = np.zeros((ROWS, COLS), dtype=np.uint8)
    for i in range(min(ROWS, COLS)):
        arr[i][i] = PIXEL_ON
        if COLS - 1 - i < COLS and i < ROWS:
            arr[i][COLS - 1 - i] = PIXEL_ON
    return Frame(arr)

def draw_frame(frame):
    """Safely draw a frame to the display."""
    try:
        Bridge.call("draw", frame.to_board_bytes(), timeout=5)
        return True
    except Exception as e:
        logger.error(f"Draw failed: {e}")
        return False

def fetch_display_state():
    """Fetch display state from API."""
    try:
        resp = requests.get(f"{API_URL}/api/status/led/display", timeout=2)
        if resp.status_code == 200:
            return resp.json()
    except Exception as e:
        logger.debug(f"API fetch: {e}")
    return None

last_mode = None
last_update = 0
last_heartbeat = 0
heartbeat_on = False

def combine_with_heartbeat(balance_arr, on=False):
    """Combine balance display with heartbeat indicator (bottom-left pixel)."""
    arr = balance_arr.copy()
    if on:
        arr[7, 0] = PIXEL_ON  # Bottom-left pixel
    return arr

def loop():
    global last_value, last_mode, scroll_offset, syncing_frame, last_update, current_balance_arr
    global last_heartbeat, heartbeat_on

    try:
        state = fetch_display_state()

        if state is None:
            draw_frame(create_error_frame())
            time.sleep(1)
            return

        mode = state.get("mode", "balance")
        value = state.get("value", 0)

        # Update balance display if needed
        if mode == "balance":
            now = time.time()
            value_changed = value != last_value or mode != last_mode or current_balance_arr is None

            # Toggle heartbeat every 2 seconds
            if now - last_heartbeat >= 2:
                heartbeat_on = not heartbeat_on
                last_heartbeat = now

            if value_changed:
                logger.info(f"Balance: {int(value/1000)}k + {int(value%1000)} EUR")
                current_balance_arr = create_balance_arr(value)
                last_value = value

            # Draw balance with heartbeat indicator
            if current_balance_arr is not None:
                combined = combine_with_heartbeat(current_balance_arr, heartbeat_on)
                draw_frame(Frame(combined))

        # Handle different modes
        if mode == "balance":
            time.sleep(1)

        elif mode == "syncing":
            if mode != last_mode:
                logger.info("Syncing...")
            draw_frame(create_syncing_frame(syncing_frame))
            syncing_frame = (syncing_frame + 1) % 20
            time.sleep(0.15)

        elif mode == "api_call":
            if mode != last_mode:
                logger.info("API call")
            draw_frame(create_syncing_frame(syncing_frame))
            syncing_frame = (syncing_frame + 1) % 20
            time.sleep(0.15)

        elif mode == "no_wifi":
            if mode != last_mode:
                logger.warning("No WiFi!")
            draw_frame(create_no_wifi_frame(scroll_offset))
            scroll_offset += 1
            time.sleep(0.15)

        else:
            draw_frame(create_balance_frame(value))
            time.sleep(1)

        last_mode = mode

    except Exception as e:
        logger.error(f"Loop error: {e}")
        time.sleep(1)

logger.info("LED Display starting (big digits + progress bar)...")
App.run(user_loop=loop)
