"""Service locator for finding and connecting to services."""

import os
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Optional, Union, cast

import yaml  # type: ignore[import-untyped]

from app.infrastructure.http_clients.coordinator_client import CoordinatorHTTPClient
from app.infrastructure.http_clients.evaluator_client import EvaluatorHTTPClient
from app.infrastructure.http_clients.gateway_client import GatewayHTTPClient
from app.infrastructure.http_clients.generator_client import GeneratorHTTPClient
from app.infrastructure.http_clients.opportunity_client import OpportunityHTTPClient
from app.infrastructure.http_clients.optimization_client import OptimizationHTTPClient
from app.infrastructure.http_clients.planning_client import PlanningHTTPClient
from app.infrastructure.http_clients.portfolio_client import PortfolioHTTPClient
from app.infrastructure.http_clients.scoring_client import ScoringHTTPClient
from app.infrastructure.http_clients.trading_client import TradingHTTPClient
from app.infrastructure.http_clients.universe_client import UniverseHTTPClient
from app.infrastructure.service_discovery.device_config import DeviceInfo


@dataclass
class ServiceConfig:
    """Service configuration."""

    name: str
    mode: str  # "local" or "remote"
    device_id: str
    port: int
    client_config: Dict[str, Any]
    health_check_config: Dict[str, Any]


@dataclass
class TLSConfig:
    """TLS/mTLS configuration."""

    enabled: bool
    mutual: bool
    ca_cert: str
    server_cert: str
    server_key: str
    client_cert: Optional[str] = None
    client_key: Optional[str] = None
    server_hostname_override: Optional[str] = None


@dataclass
class ServiceLocation:
    """Service location information."""

    name: str
    mode: str  # "local" or "remote"
    address: str  # "localhost" or IP address
    port: int
    timeout_seconds: int
    max_retries: int
    retry_backoff_ms: int


class ServiceLocator:
    """
    Locates services and provides connection information.

    Handles both local (in-process) and remote (HTTP) services.
    """

    def __init__(self, services_config_path: Optional[str] = None):
        """
        Initialize service locator.

        Args:
            services_config_path: Path to services.yaml, or None for default
        """
        services_config_path_resolved: str
        if services_config_path is None:
            app_root = Path(__file__).parent.parent.parent
            services_config_path_resolved = str(app_root / "config" / "services.yaml")
        else:
            services_config_path_resolved = services_config_path

        # Allow override via environment variable
        services_config_path_resolved = os.getenv(
            "SERVICES_CONFIG_PATH", services_config_path_resolved
        )

        with open(services_config_path_resolved, "r") as f:
            self.config = yaml.safe_load(f)

        self.deployment_mode = self.config["deployment"]["mode"]
        self.services = self.config["services"]
        self.devices = {
            dev_id: DeviceInfo(id=dev_id, address=info["address"])
            for dev_id, info in self.config["devices"].items()
        }

        # Load TLS configuration
        self.tls_config = self._load_tls_config()
        self.project_root = Path(__file__).parent.parent.parent.parent

    def _load_tls_config(self) -> Optional[TLSConfig]:
        """Load TLS configuration from services.yaml."""
        tls_conf = self.config.get("tls", {})
        if not tls_conf.get("enabled", False):
            return None

        return TLSConfig(
            enabled=tls_conf.get("enabled", False),
            mutual=tls_conf.get("mutual", False),
            ca_cert=tls_conf.get("ca_cert", "certs/ca-cert.pem"),
            server_cert=tls_conf.get("server_cert", "certs/server-cert.pem"),
            server_key=tls_conf.get("server_key", "certs/server-key.pem"),
            client_cert=tls_conf.get("client_cert"),
            client_key=tls_conf.get("client_key"),
            server_hostname_override=tls_conf.get("server_hostname_override"),
        )

    def get_service_location(self, service_name: str) -> ServiceLocation:
        """
        Get location info for a service.

        Args:
            service_name: Name of service (e.g., "planning")

        Returns:
            ServiceLocation with connection details

        Raises:
            ValueError: If service not found in config
        """
        if service_name not in self.services:
            raise ValueError(f"Service '{service_name}' not found in config")

        svc = self.services[service_name]
        mode = svc["mode"]
        device_id = svc["device_id"]
        port = svc["port"]

        # Get device address
        if mode == "local":
            address = "localhost"
        else:
            if device_id not in self.devices:
                raise ValueError(
                    f"Device '{device_id}' not found for service '{service_name}'"
                )
            address = self.devices[device_id].address

        # Get client config
        client_config = svc.get("client", {})

        return ServiceLocation(
            name=service_name,
            mode=mode,
            address=address,
            port=port,
            timeout_seconds=client_config.get("timeout_seconds", 30),
            max_retries=client_config.get("max_retries", 3),
            retry_backoff_ms=client_config.get("retry_backoff_ms", 1000),
        )

    def create_http_client(self, service_name: str) -> Union[
        UniverseHTTPClient,
        PortfolioHTTPClient,
        TradingHTTPClient,
        ScoringHTTPClient,
        OptimizationHTTPClient,
        PlanningHTTPClient,
        GatewayHTTPClient,
        OpportunityHTTPClient,
        GeneratorHTTPClient,
        EvaluatorHTTPClient,
        CoordinatorHTTPClient,
    ]:
        """
        Create HTTP client for a service.

        Args:
            service_name: Name of service

        Returns:
            HTTP client instance for the service

        Raises:
            ValueError: If service name is unknown
        """
        location = self.get_service_location(service_name)

        # Build base URL
        base_url = f"http://{location.address}:{location.port}"

        # Create appropriate client based on service name
        client_classes = {
            "universe": UniverseHTTPClient,
            "portfolio": PortfolioHTTPClient,
            "trading": TradingHTTPClient,
            "scoring": ScoringHTTPClient,
            "optimization": OptimizationHTTPClient,
            "planning": PlanningHTTPClient,
            "gateway": GatewayHTTPClient,
            "opportunity": OpportunityHTTPClient,
            "generator": GeneratorHTTPClient,
            "evaluator": EvaluatorHTTPClient,
            "coordinator": CoordinatorHTTPClient,
        }

        if service_name not in client_classes:
            raise ValueError(f"Unknown service: {service_name}")

        client_class = client_classes[service_name]
        # Type cast is safe because we validate service_name above
        return cast(
            Union[
                UniverseHTTPClient,
                PortfolioHTTPClient,
                TradingHTTPClient,
                ScoringHTTPClient,
                OptimizationHTTPClient,
                PlanningHTTPClient,
                GatewayHTTPClient,
                OpportunityHTTPClient,
                GeneratorHTTPClient,
                EvaluatorHTTPClient,
                CoordinatorHTTPClient,
            ],
            client_class(
                base_url=base_url,
                service_name=service_name,
                timeout=float(location.timeout_seconds),
            ),
        )

    def _resolve_cert_path(self, cert_path: str) -> Path:
        """
        Resolve certificate path (supports relative and absolute paths).

        Args:
            cert_path: Certificate path from config

        Returns:
            Absolute Path to certificate
        """
        path = Path(cert_path)

        # If absolute, use as-is
        if path.is_absolute():
            return path

        # Otherwise, resolve relative to project root
        return self.project_root / cert_path

    def is_service_local(self, service_name: str) -> bool:
        """Check if service runs locally (in-process)."""
        location = self.get_service_location(service_name)
        return location.mode == "local"

    def get_all_local_services(self) -> list[str]:
        """Get list of services that run locally on this device."""
        return [name for name, svc in self.services.items() if svc["mode"] == "local"]


# Global service locator instance
_service_locator: Optional[ServiceLocator] = None


def get_service_locator() -> ServiceLocator:
    """Get global service locator instance (singleton)."""
    global _service_locator
    if _service_locator is None:
        _service_locator = ServiceLocator()
    return _service_locator


def reset_service_locator():
    """Reset service locator (for testing)."""
    global _service_locator
    _service_locator = None
