"""Sentinel TUI — Neon Pulse edition."""

from __future__ import annotations

from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import VerticalScroll
from textual.widgets import Footer

from sentinel_tui.api.client import SentinelAPI
from sentinel_tui.screens.securities import SecuritiesScreen
from sentinel_tui.themes import THEME_NAMES, THEMES
from sentinel_tui.widgets import Actions, BrailleSparkline, Hero, Metrics, StatusBar


class SentinelApp(App):
    """Sentinel terminal UI — Neon Pulse."""

    TITLE = "Sentinel"
    CSS_PATH = "app.tcss"

    BINDINGS = [
        Binding("q", "quit", "Quit"),
        Binding("r", "refresh", "Refresh"),
        Binding("c", "cycle_theme", "Theme"),
        Binding("t", "toggle_table", "Table"),
    ]

    def __init__(self, api_url: str) -> None:
        super().__init__()
        self.api = SentinelAPI(api_url)
        self.api_url = api_url
        self._theme_index = 0
        self._securities: list[dict] = []

    def compose(self) -> ComposeResult:
        yield StatusBar()
        with VerticalScroll(id="main-scroll"):
            yield Hero()
            yield Metrics()
            yield BrailleSparkline()
            yield Actions()
        yield Footer()

    async def on_mount(self) -> None:
        for theme in THEMES:
            self.register_theme(theme)
        self.theme = THEME_NAMES[0]

    async def on_ready(self) -> None:
        await self._load_data()

    async def _load_data(self) -> None:
        status_bar = self.query_one(StatusBar)
        hero = self.query_one(Hero)
        metrics = self.query_one(Metrics)
        sparkline = self.query_one(BrailleSparkline)
        actions = self.query_one(Actions)

        # Health check
        try:
            health = await self.api.health()
            mode = health.get("trading_mode", "")
            status_bar.set_connection(True, mode)
        except Exception:
            status_bar.set_connection(False)
            hero.set_error(f"Cannot reach API at {self.api_url}")
            return

        # Portfolio
        try:
            portfolio = await self.api.portfolio()
            total = portfolio.get("total_value_eur", 0)
            cash = portfolio.get("total_cash_eur", 0)
            invested = total - cash
            hero.set_value(total)
        except Exception as exc:
            hero.set_error(str(exc))
            total, cash, invested = 0, 0, 0

        # P&L history
        pnl_pct = 0.0
        try:
            pnl_data = await self.api.pnl_history()
            snapshots = pnl_data.get("snapshots", [])
            pnl_values = [s.get("pnl_pct", 0) for s in snapshots]
            sparkline.set_data(pnl_values)
            if pnl_values:
                pnl_pct = pnl_values[-1]
        except Exception:
            sparkline.set_data([])

        metrics.set_data(pnl_pct, cash, invested)

        # Recommendations
        try:
            recs_data = await self.api.recommendations()
            recs = recs_data.get("recommendations", recs_data)
            if not isinstance(recs, list):
                recs = []
            actions.set_recommendations(recs)
        except Exception:
            actions.set_recommendations([])

        # Securities (cached for table screen)
        try:
            self._securities = await self.api.unified()
            self._securities.sort(
                key=lambda s: s.get("value_eur", 0) or 0, reverse=True
            )
        except Exception:
            self._securities = []

    async def action_refresh(self) -> None:
        await self._load_data()

    def action_cycle_theme(self) -> None:
        self._theme_index = (self._theme_index + 1) % len(THEME_NAMES)
        self.theme = THEME_NAMES[self._theme_index]
        self.query_one(StatusBar)._render_bar()

    def action_toggle_table(self) -> None:
        screen = SecuritiesScreen()
        self.push_screen(screen)
        if self._securities:
            screen.set_data(self._securities)

    async def on_unmount(self) -> None:
        await self.api.close()
