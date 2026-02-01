"""Top status bar widget."""

from __future__ import annotations

from textual.widgets import Static

from sentinel_tui.themes import THEME_NAMES


class StatusBar(Static):
    """Single-line header: connection dot, title, trading mode, theme index."""

    DEFAULT_CSS = """
    StatusBar {
        dock: top;
        height: 1;
        padding: 0 1;
        background: $surface;
        color: $text;
    }
    """

    def __init__(self) -> None:
        super().__init__()
        self._connected = False
        self._mode = ""

    def set_connection(self, connected: bool, mode: str = "") -> None:
        self._connected = connected
        self._mode = mode
        self._render_bar()

    def _render_bar(self) -> None:
        dot = "[$success]\u25cf[/]" if self._connected else "[$error]\u25cf[/]"
        status = "CONNECTED" if self._connected else "DISCONNECTED"

        mode_text = ""
        if self._mode:
            mode_text = f"  \u2502  {self._mode.upper()}"

        theme_idx = 1
        current = self.app.theme
        if current in THEME_NAMES:
            theme_idx = THEME_NAMES.index(current) + 1
        theme_text = f"[{theme_idx}/{len(THEME_NAMES)}]"

        self.update(f" {dot} SENTINEL  \u2502  {status}{mode_text}  \u2502  {theme_text}")
