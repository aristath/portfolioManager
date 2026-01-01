"""Centralized display update service.

Orchestrates system stats and microservice health monitoring for LED display.
"""

import logging

from app.modules.display.services.display_service import (
    set_fill_percentage,
    set_led3,
    set_led4,
)
from app.modules.display.services.service_monitor_service import (
    get_service_monitor_service,
)
from app.modules.display.services.system_stats_service import get_system_stats_service

logger = logging.getLogger(__name__)


async def update_display_stats() -> None:
    """Update display with system stats and microservice health.

    This is the single entry point for all display updates.
    Retrieves data from background services and updates display state.
    """
    try:
        # Get system stats service (auto-starts if not running)
        stats_service = get_system_stats_service()
        fill_pct = stats_service.get_fill_percentage()
        set_fill_percentage(fill_pct)

        # Get service monitor (auto-starts if not running)
        monitor_service = get_service_monitor_service()
        led3_color = monitor_service.get_led3_color()
        led4_color = monitor_service.get_led4_color()
        set_led3(*led3_color)
        set_led4(*led4_color)

        logger.debug(
            f"Display updated: fill={fill_pct:.1f}%, "
            f"LED3={led3_color}, LED4={led4_color}"
        )

    except Exception as e:
        logger.error(f"Failed to update display: {e}", exc_info=True)
        # Set safe fallback values on error
        set_fill_percentage(0.0)
        set_led3(0, 0, 0)
        set_led4(0, 0, 0)
