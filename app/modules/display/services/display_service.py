"""LED Matrix Display Service - System stats visualization.

Manages LED matrix fill percentage (based on CPU/RAM) and RGB LED states
for microservice health indication.
"""

import logging
from threading import Lock

from app.core.events import SystemEvent, emit

logger = logging.getLogger(__name__)


class DisplayStateManager:
    """Thread-safe display state manager for system stats visualization.

    Manages LED matrix fill percentage and RGB LED states for microservice health.
    Thread-safe operations ensure concurrent access from multiple jobs/API endpoints.
    """

    def __init__(self) -> None:
        """Initialize display state manager."""
        self._lock = Lock()
        self._fill_percentage: float = 0.0  # LED matrix fill percentage (0.0-100.0)
        self._led3: list[int] = [0, 0, 0]  # RGB LED 3 (microservice health)
        self._led4: list[int] = [0, 0, 0]  # RGB LED 4 (microservice health)

    def set_fill_percentage(self, percentage: float) -> None:
        """Set LED matrix fill percentage.

        Args:
            percentage: Fill percentage (0.0-100.0)
        """
        with self._lock:
            old_percentage = self._fill_percentage
            self._fill_percentage = max(0.0, min(100.0, percentage))
            if abs(old_percentage - self._fill_percentage) > 0.1:
                logger.debug(
                    f"Fill percentage updated: {old_percentage:.1f}% -> "
                    f"{self._fill_percentage:.1f}%"
                )
        emit(SystemEvent.DISPLAY_STATE_CHANGED)

    def get_fill_percentage(self) -> float:
        """Get current fill percentage.

        Returns:
            Fill percentage (0.0-100.0)
        """
        with self._lock:
            return self._fill_percentage

    def set_led3(self, r: int, g: int, b: int) -> None:
        """Set RGB LED 3 color (microservice health).

        Args:
            r: Red value (0-255)
            g: Green value (0-255)
            b: Blue value (0-255)
        """
        with self._lock:
            self._led3 = [r, g, b]
        emit(SystemEvent.DISPLAY_STATE_CHANGED)

    def get_led3(self) -> list[int]:
        """Get RGB LED 3 color.

        Returns:
            [r, g, b] values
        """
        with self._lock:
            return self._led3.copy()

    def set_led4(self, r: int, g: int, b: int) -> None:
        """Set RGB LED 4 color (microservice health).

        Args:
            r: Red value (0-255)
            g: Green value (0-255)
            b: Blue value (0-255)
        """
        with self._lock:
            self._led4 = [r, g, b]
        emit(SystemEvent.DISPLAY_STATE_CHANGED)

    def get_led4(self) -> list[int]:
        """Get RGB LED 4 color.

        Returns:
            [r, g, b] values
        """
        with self._lock:
            return self._led4.copy()


# Singleton instance for dependency injection
_display_state_manager = DisplayStateManager()


# Module-level functions for backward compatibility
def set_fill_percentage(percentage: float) -> None:
    """Set LED matrix fill percentage.

    Args:
        percentage: Fill percentage (0.0-100.0)
    """
    _display_state_manager.set_fill_percentage(percentage)


def get_fill_percentage() -> float:
    """Get current fill percentage.

    Returns:
        Fill percentage (0.0-100.0)
    """
    return _display_state_manager.get_fill_percentage()


def set_text(text: str) -> None:
    """Deprecated: Set display text (no-op).

    This function is deprecated and does nothing.
    The LED matrix now displays system stats, not text.

    Args:
        text: Ignored (kept for backwards compatibility)
    """
    # No-op for backwards compatibility
    pass


def get_current_text() -> str:
    """Deprecated: Get display text (returns empty string).

    This function is deprecated.
    The LED matrix now displays system stats, not text.

    Returns:
        Empty string (for backwards compatibility)
    """
    return ""


def set_led3(r: int, g: int, b: int) -> None:
    """Set RGB LED 3 color (microservice health).

    Args:
        r: Red value (0-255)
        g: Green value (0-255)
        b: Blue value (0-255)
    """
    _display_state_manager.set_led3(r, g, b)


def set_led4(r: int, g: int, b: int) -> None:
    """Set RGB LED 4 color (microservice health).

    Args:
        r: Red value (0-255)
        g: Green value (0-255)
        b: Blue value (0-255)
    """
    _display_state_manager.set_led4(r, g, b)
