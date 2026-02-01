"""Braille sparkline widget for P&L history."""

from __future__ import annotations

import shutil

from textual.widgets import Static

# Braille dot positions map to vertical offsets within a 2x4 cell.
# We use a simplified 1-column approach: each braille char encodes 4 vertical dots.
# Dots 1,2,3,7 map to rows 0-3 from top to bottom in the left column.
_DOT_OFFSETS = [0x01, 0x02, 0x04, 0x40]  # rows 0, 1, 2, 3


def _braille_sparkline(values: list[float], width: int) -> str:
    """Encode a list of floats as a braille sparkline string.

    Each character represents one data point, using braille dots to show
    the value height (0-3 dots filled from bottom).
    """
    if not values or width <= 0:
        return ""

    # Resample to fit width
    resampled: list[float] = []
    step = len(values) / width
    for i in range(width):
        idx = int(i * step)
        resampled.append(values[min(idx, len(values) - 1)])

    # Normalize to 0-3 range
    lo = min(resampled)
    hi = max(resampled)
    spread = hi - lo if hi != lo else 1.0
    normalized = [int((v - lo) / spread * 3 + 0.5) for v in resampled]

    # Encode as braille: fill dots from bottom up
    chars: list[str] = []
    for level in normalized:
        code = 0x2800  # empty braille
        # Fill bottom-up: level 0 = empty, 1 = bottom dot, 2 = bottom 2, etc.
        for dot in range(level + 1):
            row = 3 - dot  # bottom-up: row 3, 2, 1, 0
            if row >= 0:
                code |= _DOT_OFFSETS[row]
        chars.append(chr(code))

    return "".join(chars)


class BrailleSparkline(Static):
    """Renders a braille sparkline of P&L history."""

    DEFAULT_CSS = """
    BrailleSparkline {
        height: auto;
        padding: 0 2;
        text-align: center;
        content-align: center middle;
    }
    """

    def set_data(self, values: list[float]) -> None:
        if not values:
            self.update("")
            return
        cols = shutil.get_terminal_size().columns
        width = min(len(values), cols - 4)
        line = _braille_sparkline(values, width)
        self.update(f"[$primary]{line}[/]")
