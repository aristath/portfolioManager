"""Neon Pulse theme definitions."""

from textual.theme import Theme

THEMES = [
    Theme(
        name="synthwave-sunset",
        primary="#ff00ff",
        secondary="#00ffff",
        accent="#ff00ff",
        background="#1a0b2e",
        surface="#2d1b4e",
        panel="#2d1b4e",
        error="#ff4444",
        warning="#ffaa00",
        success="#00ff88",
        dark=True,
    ),
    Theme(
        name="cyberpunk-matrix",
        primary="#00ff41",
        secondary="#ff00ff",
        accent="#00ff41",
        background="#0a0a0a",
        surface="#1a1a1a",
        panel="#1a1a1a",
        error="#ff4444",
        warning="#ffaa00",
        success="#00ff41",
        dark=True,
    ),
    Theme(
        name="neon-tron",
        primary="#00f3ff",
        secondary="#ff9d00",
        accent="#00f3ff",
        background="#0d1b2a",
        surface="#1b2d45",
        panel="#1b2d45",
        error="#ff4444",
        warning="#ff9d00",
        success="#00ff88",
        dark=True,
    ),
    Theme(
        name="cyberpunk-night",
        primary="#ff0099",
        secondary="#00d4ff",
        accent="#ff0099",
        background="#1a1a2e",
        surface="#2a2a4e",
        panel="#2a2a4e",
        error="#ff4444",
        warning="#ffaa00",
        success="#00ff88",
        dark=True,
    ),
]

THEME_NAMES = [t.name for t in THEMES]
