"""
Sentinel LED Display - Python App

Arduino App Lab application for controlling the LED display on Arduino Uno Q.
Exposes REST API endpoints via Web UI Brick for external communication.

Following Arduino Uno Q documentation:
https://docs.arduino.cc/tutorials/uno-q/user-manual/

Endpoints:
- POST /text - Scroll text across LED matrix
- POST /led3 - Set RGB LED 3 color (sync indicator)
- POST /led4 - Set RGB LED 4 color (processing indicator)
- POST /clear - Clear the LED matrix
- POST /pixels - Set pixel count (for system stats mode)
- POST /draw - Draw raw bitmap to matrix
- GET /health - Health check endpoint
"""

from arduino.app_utils import App, Bridge
from arduino.app_bricks.web_ui import WebUI
from fastapi import Request
import threading
import time

# Initialize Web UI Brick for REST API
ui = WebUI()

# Matrix dimensions
MATRIX_ROWS = 8
MATRIX_COLS = 13

# Scrolling text state
scroll_thread = None
scroll_stop_event = threading.Event()

# 5x7 bitmap font for uppercase letters, numbers, and common symbols
# Each character is 5 columns wide, stored as column bytes (LSB = top row)
FONT_5X7 = {
    ' ': [0x00, 0x00, 0x00, 0x00, 0x00],
    'A': [0x7E, 0x09, 0x09, 0x09, 0x7E],
    'B': [0x7F, 0x49, 0x49, 0x49, 0x36],
    'C': [0x3E, 0x41, 0x41, 0x41, 0x22],
    'D': [0x7F, 0x41, 0x41, 0x41, 0x3E],
    'E': [0x7F, 0x49, 0x49, 0x49, 0x41],
    'F': [0x7F, 0x09, 0x09, 0x09, 0x01],
    'G': [0x3E, 0x41, 0x49, 0x49, 0x7A],
    'H': [0x7F, 0x08, 0x08, 0x08, 0x7F],
    'I': [0x00, 0x41, 0x7F, 0x41, 0x00],
    'J': [0x20, 0x40, 0x41, 0x3F, 0x01],
    'K': [0x7F, 0x08, 0x14, 0x22, 0x41],
    'L': [0x7F, 0x40, 0x40, 0x40, 0x40],
    'M': [0x7F, 0x02, 0x0C, 0x02, 0x7F],
    'N': [0x7F, 0x04, 0x08, 0x10, 0x7F],
    'O': [0x3E, 0x41, 0x41, 0x41, 0x3E],
    'P': [0x7F, 0x09, 0x09, 0x09, 0x06],
    'Q': [0x3E, 0x41, 0x51, 0x21, 0x5E],
    'R': [0x7F, 0x09, 0x19, 0x29, 0x46],
    'S': [0x46, 0x49, 0x49, 0x49, 0x31],
    'T': [0x01, 0x01, 0x7F, 0x01, 0x01],
    'U': [0x3F, 0x40, 0x40, 0x40, 0x3F],
    'V': [0x1F, 0x20, 0x40, 0x20, 0x1F],
    'W': [0x3F, 0x40, 0x38, 0x40, 0x3F],
    'X': [0x63, 0x14, 0x08, 0x14, 0x63],
    'Y': [0x07, 0x08, 0x70, 0x08, 0x07],
    'Z': [0x61, 0x51, 0x49, 0x45, 0x43],
    '0': [0x3E, 0x51, 0x49, 0x45, 0x3E],
    '1': [0x00, 0x42, 0x7F, 0x40, 0x00],
    '2': [0x42, 0x61, 0x51, 0x49, 0x46],
    '3': [0x21, 0x41, 0x45, 0x4B, 0x31],
    '4': [0x18, 0x14, 0x12, 0x7F, 0x10],
    '5': [0x27, 0x45, 0x45, 0x45, 0x39],
    '6': [0x3C, 0x4A, 0x49, 0x49, 0x30],
    '7': [0x01, 0x71, 0x09, 0x05, 0x03],
    '8': [0x36, 0x49, 0x49, 0x49, 0x36],
    '9': [0x06, 0x49, 0x49, 0x29, 0x1E],
    '.': [0x00, 0x60, 0x60, 0x00, 0x00],
    ',': [0x00, 0x80, 0x60, 0x00, 0x00],
    ':': [0x00, 0x36, 0x36, 0x00, 0x00],
    '-': [0x08, 0x08, 0x08, 0x08, 0x08],
    '+': [0x08, 0x08, 0x3E, 0x08, 0x08],
    '*': [0x14, 0x08, 0x3E, 0x08, 0x14],
    '/': [0x20, 0x10, 0x08, 0x04, 0x02],
    '$': [0x24, 0x2A, 0x7F, 0x2A, 0x12],
    '%': [0x23, 0x13, 0x08, 0x64, 0x62],
    '(': [0x00, 0x1C, 0x22, 0x41, 0x00],
    ')': [0x00, 0x41, 0x22, 0x1C, 0x00],
    '!': [0x00, 0x00, 0x5F, 0x00, 0x00],
    '?': [0x02, 0x01, 0x51, 0x09, 0x06],
    '=': [0x14, 0x14, 0x14, 0x14, 0x14],
    '_': [0x40, 0x40, 0x40, 0x40, 0x40],
}


def render_text_to_bitmap(text):
    """Render text string to a wide bitmap buffer (list of column bytes)."""
    bitmap = []
    for char in text.upper():
        if char in FONT_5X7:
            bitmap.extend(FONT_5X7[char])
        else:
            bitmap.extend(FONT_5X7[' '])  # Unknown char = space
        bitmap.append(0x00)  # 1 column gap between characters
    return bitmap


def extract_frame(bitmap, offset, width=MATRIX_COLS):
    """Extract a frame from the bitmap at given offset."""
    frame = []
    for col in range(width):
        idx = offset + col
        if 0 <= idx < len(bitmap):
            frame.append(bitmap[idx])
        else:
            frame.append(0x00)
    return frame


def columns_to_bitmap(columns):
    """Convert column-based font data to 8x13 bitmap array.
    
    Returns 104 bytes in row-major order.
    Each byte is brightness level 0-7 (3-bit grayscale per Uno Q docs).
    0 = off, 7 = full brightness.
    Font columns are stored with LSB = top row (row 0).
    """
    bitmap = []
    for row in range(MATRIX_ROWS):
        for col in range(MATRIX_COLS):
            if col < len(columns):
                # Check if this pixel is set in the column byte
                # Bit 0 = row 0 (top), bit 7 = row 7 (bottom)
                # 7 = on (full brightness), 0 = off
                pixel = 7 if (columns[col] & (1 << row)) else 0
            else:
                pixel = 0  # Background = off
            bitmap.append(pixel)
    return bitmap


def scroll_text_loop(text, speed_ms):
    """Background thread function to scroll text across the matrix."""
    global scroll_stop_event
    
    # Render full text to column-based bitmap
    text_columns = render_text_to_bitmap(text)
    
    # Add padding at start and end for smooth scroll
    padding = [0x00] * MATRIX_COLS
    text_columns = padding + text_columns + padding
    
    total_width = len(text_columns)
    offset = 0
    
    while not scroll_stop_event.is_set():
        # Extract current frame (13 columns)
        frame_cols = extract_frame(text_columns, offset, MATRIX_COLS)
        
        # Convert to 104-byte bitmap
        frame_bytes = columns_to_bitmap(frame_cols)
        
        # Send to MCU
        try:
            Bridge.call("drawMatrix", bytes(frame_bytes))
        except Exception:
            pass  # Ignore errors during scroll
        
        # Advance scroll position
        offset += 1
        if offset >= total_width - MATRIX_COLS:
            offset = 0  # Loop back
        
        # Wait for next frame
        scroll_stop_event.wait(speed_ms / 1000.0)


def start_scroll(text, speed_ms):
    """Start scrolling text (stops any existing scroll)."""
    global scroll_thread, scroll_stop_event
    
    # Stop existing scroll
    if scroll_thread is not None and scroll_thread.is_alive():
        scroll_stop_event.set()
        scroll_thread.join(timeout=1.0)
    
    # Reset stop event and start new thread
    scroll_stop_event = threading.Event()
    scroll_thread = threading.Thread(
        target=scroll_text_loop,
        args=(text, speed_ms),
        daemon=True
    )
    scroll_thread.start()


async def handle_set_text(request: Request):
    """Handle text scroll request."""
    data = await request.json()
    text = str(data.get("text", ""))
    speed = int(data.get("speed", 50))  # ms per scroll step
    
    if text:
        start_scroll(text, speed)
    else:
        # Empty text = stop scrolling and clear
        global scroll_stop_event
        scroll_stop_event.set()
        Bridge.call("clearMatrix")
    
    return {"status": "ok", "text": text, "speed": speed}


async def handle_set_led3(request: Request):
    data = await request.json()
    r = max(0, min(255, int(data.get("r", 0))))
    g = max(0, min(255, int(data.get("g", 0))))
    b = max(0, min(255, int(data.get("b", 0))))
    Bridge.call("setRGB3", r, g, b)
    return {"status": "ok", "r": r, "g": g, "b": b}


async def handle_set_led4(request: Request):
    data = await request.json()
    r = max(0, min(255, int(data.get("r", 0))))
    g = max(0, min(255, int(data.get("g", 0))))
    b = max(0, min(255, int(data.get("b", 0))))
    Bridge.call("setRGB4", r, g, b)
    return {"status": "ok", "r": r, "g": g, "b": b}


async def handle_clear_matrix(request: Request):
    global scroll_stop_event
    scroll_stop_event.set()  # Stop any scrolling
    Bridge.call("clearMatrix")
    return {"status": "ok"}


async def handle_set_pixels(request: Request):
    data = await request.json()
    count = max(0, min(104, int(data.get("count", 0))))
    Bridge.call("setPixelCount", count)
    return {"status": "ok", "count": count}


async def handle_draw_matrix(request: Request):
    data = await request.json()
    # Expect a list of bytes representing the matrix state
    pixels = data.get("pixels", [])
    if pixels:
        Bridge.call("drawMatrix", bytes(pixels))
    return {"status": "ok"}


def handle_health():
    return {"status": "healthy", "service": "sentinel-display"}


ui.expose_api("POST", "/text", handle_set_text)
ui.expose_api("POST", "/led3", handle_set_led3)
ui.expose_api("POST", "/led4", handle_set_led4)
ui.expose_api("POST", "/clear", handle_clear_matrix)
ui.expose_api("POST", "/pixels", handle_set_pixels)
ui.expose_api("POST", "/draw", handle_draw_matrix)
ui.expose_api("GET", "/health", handle_health)


if __name__ == "__main__":
    App.run()
