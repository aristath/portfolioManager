"""Hero portfolio value widget — large figlet text."""

from __future__ import annotations

import shutil

import pyfiglet
from textual.widgets import Static


class Hero(Static):
    """Displays portfolio total value in large ansi_regular figlet font."""

    DEFAULT_CSS = """
    Hero {
        height: auto;
        padding: 1 2;
        text-align: center;
        content-align: center middle;
    }
    """

    def set_value(self, value: float, currency: str = "\u20ac") -> None:
        formatted = f"{currency}{value:,.0f}"
        cols = shutil.get_terminal_size().columns
        if cols >= 60:
            figlet_text = pyfiglet.figlet_format(formatted, font="ansi_regular")
            # Strip trailing blank lines but keep structure
            lines = figlet_text.rstrip("\n").split("\n")
            figlet_text = "\n".join(lines)
            self.update(f"[$primary]{figlet_text}[/]")
        else:
            self.update(f"[$primary]{formatted}[/]")

    def set_error(self, msg: str) -> None:
        self.update(f"[$error]{msg}[/]")
