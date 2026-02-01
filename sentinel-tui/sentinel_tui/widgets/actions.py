"""Trade recommendations widget — double_blocky figlet actions."""

from __future__ import annotations

import pyfiglet
from textual.widgets import Static


class Actions(Static):
    """Displays top planner recommendations in large figlet text."""

    DEFAULT_CSS = """
    Actions {
        height: auto;
        padding: 1 2;
        border-top: solid $primary;
    }
    """

    def set_recommendations(self, recs: list[dict]) -> None:
        if not recs:
            self.update("[$secondary]No recommendations[/]")
            return

        lines: list[str] = []
        lines.append("[$primary]\u2500\u2500\u2500 NEXT ACTIONS \u2500\u2500\u2500[/]")
        lines.append("")

        buy_total = 0.0
        sell_total = 0.0

        for rec in recs[:3]:
            action = rec.get("action", "").upper()
            symbol = rec.get("symbol", "")
            cost = rec.get("price", 0) * rec.get("quantity", 0)
            text = f"{action} E{cost:,.0f} {symbol}"

            figlet_text = pyfiglet.figlet_format(text, font="double_blocky")
            fig_lines = figlet_text.rstrip("\n").split("\n")
            fig_lines = [l for l in fig_lines if l.strip()]

            reason = rec.get("reason", "")
            color = "$success" if action == "BUY" else "$secondary"

            if action == "BUY":
                buy_total += cost
            else:
                sell_total += cost

            # Place reason next to the second figlet line
            for i, fl in enumerate(fig_lines):
                if i == 1 and reason:
                    lines.append(f"[{color}]{fl}[/]  [dim]{reason}[/]")
                else:
                    lines.append(f"[{color}]{fl}[/]")

            lines.append("")

        # Summary
        parts: list[str] = []
        if buy_total:
            parts.append(f"[$success]BUY \u20ac{buy_total:,.0f}[/]")
        if sell_total:
            parts.append(f"[$secondary]SELL \u20ac{sell_total:,.0f}[/]")
        if parts:
            lines.append("  ".join(parts))

        self.update("\n".join(lines))
