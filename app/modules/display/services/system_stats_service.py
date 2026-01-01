"""System statistics collection service.

Collects CPU and memory usage statistics and calculates fill percentage
for LED matrix display visualization.
"""

import logging
import threading
import time
from typing import Optional

logger = logging.getLogger(__name__)

# Optional psutil import - gracefully degrade if not available
try:
    import psutil

    PSUTIL_AVAILABLE = True
except ImportError:
    PSUTIL_AVAILABLE = False
    logger.warning("psutil not available, system stats will use /proc fallback")


class SystemStatsService:
    """Service for collecting system statistics.

    Polls CPU and memory usage at 0.5Hz (every 2000ms) and calculates
    a fill percentage for the LED matrix display.
    """

    def __init__(self) -> None:
        """Initialize system stats service."""
        self._fill_percentage: float = 0.0
        self._running: bool = False
        self._thread: Optional[threading.Thread] = None
        self._lock = threading.Lock()

    def start(self) -> None:
        """Start the stats collection thread."""
        if self._running:
            logger.warning("SystemStatsService already running")
            return

        self._running = True
        self._thread = threading.Thread(target=self._poll_stats, daemon=True)
        self._thread.start()
        logger.info("SystemStatsService started (polling at 0.5Hz)")

    def stop(self) -> None:
        """Stop the stats collection thread."""
        self._running = False
        if self._thread:
            self._thread.join(timeout=1.0)
        logger.info("SystemStatsService stopped")

    def get_fill_percentage(self) -> float:
        """Get current fill percentage for LED matrix.

        Returns:
            Fill percentage (0.0-100.0)
        """
        with self._lock:
            return self._fill_percentage

    def _poll_stats(self) -> None:
        """Background thread that polls system stats at 0.5Hz."""
        while self._running:
            try:
                cpu_percent = self._get_cpu_percent()
                memory_percent = self._get_memory_percent()

                # Calculate fill percentage as average of CPU and RAM
                fill_percentage = (cpu_percent + memory_percent) / 2.0

                with self._lock:
                    self._fill_percentage = fill_percentage

                logger.debug(
                    f"System stats: CPU={cpu_percent:.1f}% MEM={memory_percent:.1f}% "
                    f"FILL={fill_percentage:.1f}%"
                )

            except Exception as e:
                logger.error(f"Error collecting system stats: {e}", exc_info=True)
                # Don't crash the thread on error
                with self._lock:
                    self._fill_percentage = 0.0

            # Sleep for 2000ms (0.5Hz polling rate)
            time.sleep(2.0)

    def _get_cpu_percent(self) -> float:
        """Get CPU usage percentage.

        Returns:
            CPU percentage (0.0-100.0)
        """
        if PSUTIL_AVAILABLE:
            # Use psutil for accurate CPU measurement
            return psutil.cpu_percent(interval=None)
        else:
            # Fallback to /proc/stat parsing
            return self._get_cpu_percent_from_proc()

    def _get_memory_percent(self) -> float:
        """Get memory usage percentage.

        Returns:
            Memory percentage (0.0-100.0)
        """
        if PSUTIL_AVAILABLE:
            # Use psutil for accurate memory measurement
            return psutil.virtual_memory().percent
        else:
            # Fallback to /proc/meminfo parsing
            return self._get_memory_percent_from_proc()

    def _get_cpu_percent_from_proc(self) -> float:
        """Parse /proc/stat to get CPU usage.

        Returns:
            CPU percentage (0.0-100.0)
        """
        try:
            with open("/proc/stat", "r") as f:
                line = f.readline()
                fields = line.split()
                # Fields: cpu user nice system idle iowait irq softirq
                idle = int(fields[4])
                total = sum(int(x) for x in fields[1:8])

                # Calculate percentage (this is a simplified version)
                # For accurate measurement, need to track previous values
                if total > 0:
                    return 100.0 - (idle / total * 100.0)
                return 0.0
        except Exception as e:
            logger.warning(f"Failed to read CPU from /proc/stat: {e}")
            return 0.0

    def _get_memory_percent_from_proc(self) -> float:
        """Parse /proc/meminfo to get memory usage.

        Returns:
            Memory percentage (0.0-100.0)
        """
        try:
            mem_total = 0
            mem_available = 0

            with open("/proc/meminfo", "r") as f:
                for line in f:
                    if line.startswith("MemTotal:"):
                        mem_total = int(line.split()[1])
                    elif line.startswith("MemAvailable:"):
                        mem_available = int(line.split()[1])

            if mem_total > 0:
                mem_used = mem_total - mem_available
                return (mem_used / mem_total) * 100.0
            return 0.0
        except Exception as e:
            logger.warning(f"Failed to read memory from /proc/meminfo: {e}")
            return 0.0


# Global singleton instance
_system_stats_service: Optional[SystemStatsService] = None


def get_system_stats_service() -> SystemStatsService:
    """Get or create the system stats service singleton.

    Returns:
        SystemStatsService instance
    """
    global _system_stats_service
    if _system_stats_service is None:
        _system_stats_service = SystemStatsService()
        _system_stats_service.start()
    return _system_stats_service
