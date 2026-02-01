"""Metrics row: P/L%, cash, invested."""

from __future__ import annotations

from textual.widgets import Static


class Metrics(Static):
    """Single-line summary of key portfolio metrics."""

    DEFAULT_CSS = """
    Metrics {
        height: auto;
        padding: 0 2;
        text-align: center;
        content-align: center middle;
    }
    """

    def set_data(
        self,
        pnl_pct: float,
        cash: float,
        invested: float,
        currency: str = "\u20ac",
    ) -> None:
        arrow = "\u25b2" if pnl_pct >= 0 else "\u25bc"
        color = "$success" if pnl_pct >= 0 else "$error"
        self.update(
            f"[{color}]{arrow} {pnl_pct:+.2f}%[/]"
            f"  \u2502  Cash: {currency}{cash:,.0f}"
            f"  \u2502  Invested: {currency}{invested:,.0f}"
        )
