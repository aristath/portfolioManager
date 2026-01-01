"""Microservice health monitoring service.

Monitors health of all microservices and assigns RGB colors to LED3 and LED4
based on which services are running.
"""

import asyncio
import logging
import random
import threading
from pathlib import Path
from typing import Optional

import httpx
import yaml  # type: ignore

logger = logging.getLogger(__name__)


class ServiceMonitorService:
    """Service for monitoring microservice health and LED color assignment.

    Polls all services at 1Hz to check health status and updates RGB LED colors
    based on which services are running. Services are split between LED3 and LED4
    by index (odd -> LED3, even -> LED4).
    """

    def __init__(self) -> None:
        """Initialize service monitor."""
        self._running: bool = False
        self._thread: Optional[threading.Thread] = None
        self._lock = threading.Lock()
        self._led3_color: list[int] = [0, 0, 0]
        self._led4_color: list[int] = [0, 0, 0]

        # Service discovery from services.yaml
        self._services: list[dict] = []
        self._service_colors: dict[str, tuple[int, int, int]] = {}

    def start(self) -> None:
        """Start the service monitoring thread."""
        if self._running:
            logger.warning("ServiceMonitorService already running")
            return

        # Discover services from config
        self._discover_services()

        self._running = True
        self._thread = threading.Thread(target=self._monitor_loop, daemon=True)
        self._thread.start()
        logger.info(
            f"ServiceMonitorService started (monitoring {len(self._services)} services)"
        )

    def stop(self) -> None:
        """Stop the service monitoring thread."""
        self._running = False
        if self._thread:
            self._thread.join(timeout=2.0)
        logger.info("ServiceMonitorService stopped")

    def get_led3_color(self) -> list[int]:
        """Get RGB color for LED3.

        Returns:
            [r, g, b] color values (0-255)
        """
        with self._lock:
            return self._led3_color.copy()

    def get_led4_color(self) -> list[int]:
        """Get RGB color for LED4.

        Returns:
            [r, g, b] color values (0-255)
        """
        with self._lock:
            return self._led4_color.copy()

    def _discover_services(self) -> None:
        """Discover services from services.yaml configuration."""
        try:
            # Read services.yaml from app/config directory
            config_path = (
                Path(__file__).parent.parent.parent.parent / "config" / "services.yaml"
            )
            with open(config_path, "r") as f:
                config = yaml.safe_load(f)
            services_config = config.get("services", {})

            self._services = []
            for service_name, service_config in services_config.items():
                # Handle services with multiple instances (like evaluator)
                if "instances" in service_config:
                    for idx, instance in enumerate(service_config["instances"]):
                        instance_name = f"{service_name}-{idx+1}"
                        port = instance.get("port")
                        health_endpoint = instance.get("health_endpoint", "/health")
                        self._services.append(
                            {
                                "name": instance_name,
                                "port": port,
                                "health_endpoint": health_endpoint,
                            }
                        )
                else:
                    # Single instance service
                    # Use http_port for gateway, otherwise use port
                    port = service_config.get("http_port") or service_config.get("port")
                    health_endpoint = service_config.get("health_endpoint", "/health")
                    self._services.append(
                        {
                            "name": service_name,
                            "port": port,
                            "health_endpoint": health_endpoint,
                        }
                    )

            # Assign deterministic random colors to each service
            for service in self._services:
                self._service_colors[service["name"]] = self._get_service_color(
                    service["name"]
                )

            logger.info(
                f"Discovered {len(self._services)} service instances: "
                f"{[s['name'] for s in self._services]}"
            )

        except Exception as e:
            logger.error(f"Failed to discover services: {e}", exc_info=True)
            self._services = []

    def _get_service_color(self, service_name: str) -> tuple[int, int, int]:
        """Get deterministic random color for a service.

        Args:
            service_name: Name of the service

        Returns:
            RGB color tuple (r, g, b) with values 50-255
        """
        # Use hash of service name as seed for deterministic randomness
        random.seed(hash(service_name))
        color = (
            random.randint(50, 255),
            random.randint(50, 255),
            random.randint(50, 255),
        )
        # Reset random seed to avoid affecting other random operations
        random.seed()
        return color

    def _monitor_loop(self) -> None:
        """Background thread that monitors services at 1Hz."""
        # Create event loop for this thread
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)

        while self._running:
            try:
                # Check all services
                loop.run_until_complete(self._check_all_services())

            except Exception as e:
                logger.error(f"Error in service monitor loop: {e}", exc_info=True)

            # Sleep for 1 second (1Hz polling rate)
            for _ in range(10):  # Check _running every 100ms for faster shutdown
                if not self._running:
                    break
                asyncio.run(asyncio.sleep(0.1))

        loop.close()

    async def _check_all_services(self) -> None:
        """Check health of all services and update LED colors."""
        # Check all services concurrently
        tasks = [self._check_service_health(service) for service in self._services]
        health_results = await asyncio.gather(*tasks, return_exceptions=True)

        # Separate services by LED assignment (odd index -> LED3, even -> LED4)
        led3_services = []
        led4_services = []

        for idx, (service, is_healthy) in enumerate(
            zip(self._services, health_results)
        ):
            if isinstance(is_healthy, Exception):
                logger.debug(
                    f"Service {service['name']} health check failed: {is_healthy}"
                )
                is_healthy = False

            if is_healthy:
                # Assign to LED based on index
                if idx % 2 == 0:  # Even index -> LED4
                    led4_services.append(service["name"])
                else:  # Odd index -> LED3
                    led3_services.append(service["name"])

        # Calculate blended colors for each LED
        led3_color = self._blend_service_colors(led3_services)
        led4_color = self._blend_service_colors(led4_services)

        # Update LED colors
        with self._lock:
            self._led3_color = led3_color
            self._led4_color = led4_color

        logger.debug(
            f"Services: LED3={led3_services} ({led3_color}), "
            f"LED4={led4_services} ({led4_color})"
        )

    async def _check_service_health(self, service: dict) -> bool:
        """Check if a service is healthy.

        Args:
            service: Service configuration dict

        Returns:
            True if service is healthy, False otherwise
        """
        try:
            port = service["port"]
            health_endpoint = service["health_endpoint"]
            url = f"http://localhost:{port}{health_endpoint}"

            async with httpx.AsyncClient() as client:
                response = await client.get(url, timeout=0.5)
                return response.status_code == 200

        except Exception as e:
            logger.debug(f"Service {service['name']} health check failed: {e}")
            return False

    def _blend_service_colors(self, service_names: list[str]) -> list[int]:
        """Blend RGB colors of multiple services.

        Args:
            service_names: List of service names

        Returns:
            Blended RGB color [r, g, b]
        """
        if not service_names:
            return [0, 0, 0]  # All OFF if no services running

        # Get colors for all running services
        colors = [self._service_colors[name] for name in service_names]

        # Calculate average RGB
        r = int(sum(c[0] for c in colors) / len(colors))
        g = int(sum(c[1] for c in colors) / len(colors))
        b = int(sum(c[2] for c in colors) / len(colors))

        return [r, g, b]


# Global singleton instance
_service_monitor_service: Optional[ServiceMonitorService] = None


def get_service_monitor_service() -> ServiceMonitorService:
    """Get or create the service monitor singleton.

    Returns:
        ServiceMonitorService instance
    """
    global _service_monitor_service
    if _service_monitor_service is None:
        _service_monitor_service = ServiceMonitorService()
        _service_monitor_service.start()
    return _service_monitor_service
