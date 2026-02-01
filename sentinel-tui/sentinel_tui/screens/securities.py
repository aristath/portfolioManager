"""Securities table screen — full overlay with DataTable."""

from __future__ import annotations

from rich.text import Text
from textual.app import ComposeResult
from textual.binding import Binding
from textual.screen import Screen
from textual.widgets import DataTable, Footer, Static


class SecuritiesScreen(Screen):
    """Full-screen securities table, pushed as overlay."""

    BINDINGS = [
        Binding("escape", "dismiss", "Back"),
        Binding("t", "dismiss", "Back"),
    ]

    CSS = """
    SecuritiesScreen {
        background: $background;
    }
    #sec-header {
        dock: top;
        height: 1;
        padding: 0 1;
        background: $surface;
        color: $text;
    }
    #sec-table {
        height: 1fr;
    }
    """

    def compose(self) -> ComposeResult:
        yield Static(
            " [$primary]\u25c0[/] SECURITIES  [dim]ESC to close[/]",
            id="sec-header",
        )
        yield DataTable(id="sec-table")
        yield Footer()

    def on_mount(self) -> None:
        table = self.query_one("#sec-table", DataTable)
        table.add_columns("Symbol", "Name", "Value (\u20ac)", "P/L (%)", "Score")

    def set_data(self, securities: list[dict]) -> None:
        table = self.query_one("#sec-table", DataTable)
        table.clear()
        success = self.app.current_theme.success or "#00ff88"
        error = self.app.current_theme.error or "#ff4444"
        for sec in securities:
            symbol = sec.get("symbol", "")
            name = sec.get("name", "")
            value = sec.get("value_eur", 0) or 0
            has_pos = sec.get("has_position", False)
            score = sec.get("planner_score", 0) or 0

            if has_pos:
                value_text = f"{value:,.2f}"
                pct = sec.get("profit_pct", 0) or 0
                color = success if pct >= 0 else error
                pl_text = Text(f"{pct:+.2f}%", style=color)
            else:
                value_text = Text("-", style="dim")
                pl_text = Text("-", style="dim")

            score_text = f"{score:.1f}" if score else Text("-", style="dim")
            table.add_row(symbol, name, value_text, pl_text, score_text)

    def action_dismiss(self) -> None:
        self.app.pop_screen()
