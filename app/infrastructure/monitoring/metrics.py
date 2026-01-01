"""Prometheus metrics for monitoring microservices."""

import logging
import time
from dataclasses import dataclass, field
from typing import Dict

logger = logging.getLogger(__name__)


@dataclass
class Counter:
    """Simple counter metric."""

    value: int = 0
    labels: Dict[str, str] = field(default_factory=dict)

    def inc(self, amount: int = 1):
        """Increment counter."""
        self.value += amount

    def reset(self):
        """Reset counter to zero."""
        self.value = 0


@dataclass
class Gauge:
    """Simple gauge metric."""

    value: float = 0.0
    labels: Dict[str, str] = field(default_factory=dict)

    def set(self, value: float):
        """Set gauge value."""
        self.value = value

    def inc(self, amount: float = 1.0):
        """Increment gauge."""
        self.value += amount

    def dec(self, amount: float = 1.0):
        """Decrement gauge."""
        self.value -= amount


@dataclass
class HistogramBucket:
    """Histogram bucket."""

    le: float  # Upper bound (less than or equal)
    count: int = 0


class Histogram:
    """Simple histogram metric for tracking distributions."""

    def __init__(
        self,
        buckets: list[float] | None = None,
        labels: Dict[str, str] | None = None,
    ):
        """Initialize histogram."""
        self.labels = labels or {}

        # Default buckets if none provided
        if buckets is None:
            buckets = [
                0.005,
                0.01,
                0.025,
                0.05,
                0.1,
                0.25,
                0.5,
                1.0,
                2.5,
                5.0,
                10.0,
                float("inf"),
            ]

        self.buckets = [HistogramBucket(le=b) for b in sorted(buckets)]
        self.sum = 0.0
        self.count = 0

    def observe(self, value: float):
        """Record an observation."""
        self.sum += value
        self.count += 1

        for bucket in self.buckets:
            if value <= bucket.le:
                bucket.count += 1

    def reset(self):
        """Reset histogram."""
        for bucket in self.buckets:
            bucket.count = 0
        self.sum = 0.0
        self.count = 0


class MetricsRegistry:
    """Registry for managing metrics."""

    def __init__(self):
        """Initialize registry."""
        self._counters: Dict[str, Counter] = {}
        self._gauges: Dict[str, Gauge] = {}
        self._histograms: Dict[str, Histogram] = {}

    def counter(self, name: str, labels: Dict[str, str] | None = None) -> Counter:
        """Get or create a counter."""
        key = self._make_key(name, labels)
        if key not in self._counters:
            self._counters[key] = Counter(labels=labels or {})
        return self._counters[key]

    def gauge(self, name: str, labels: Dict[str, str] | None = None) -> Gauge:
        """Get or create a gauge."""
        key = self._make_key(name, labels)
        if key not in self._gauges:
            self._gauges[key] = Gauge(labels=labels or {})
        return self._gauges[key]

    def histogram(
        self,
        name: str,
        buckets: list[float] | None = None,
        labels: Dict[str, str] | None = None,
    ) -> Histogram:
        """Get or create a histogram."""
        key = self._make_key(name, labels)
        if key not in self._histograms:
            self._histograms[key] = Histogram(buckets=buckets, labels=labels or {})
        return self._histograms[key]

    def _make_key(self, name: str, labels: Dict[str, str] | None) -> str:
        """Create unique key for metric with labels."""
        if not labels:
            return name
        label_str = ",".join(f"{k}={v}" for k, v in sorted(labels.items()))
        return f"{name}{{{label_str}}}"

    def export_text(self) -> str:
        """Export metrics in Prometheus text format."""
        lines = []

        # Export counters
        for key, counter in self._counters.items():
            name = key.split("{")[0]
            lines.append(f"# TYPE {name} counter")
            lines.append(f"{key} {counter.value}")

        # Export gauges
        for key, gauge in self._gauges.items():
            name = key.split("{")[0]
            lines.append(f"# TYPE {name} gauge")
            lines.append(f"{key} {gauge.value}")

        # Export histograms
        for key, histogram in self._histograms.items():
            name = key.split("{")[0]
            lines.append(f"# TYPE {name} histogram")

            # Export buckets
            for bucket in histogram.buckets:
                bucket_key = (
                    f'{key}_bucket{{le="{bucket.le}"}}'
                    if "{" not in key
                    else key.replace("}", f',le="{bucket.le}"}}"')
                )
                lines.append(f"{bucket_key} {bucket.count}")

            # Export sum and count
            sum_key = f"{key}_sum" if "{" not in key else key.replace("}", "_sum}")
            count_key = (
                f"{key}_count" if "{" not in key else key.replace("}", "_count}")
            )
            lines.append(f"{sum_key} {histogram.sum}")
            lines.append(f"{count_key} {histogram.count}")

        return "\n".join(lines) + "\n"


# Global registry
_registry = MetricsRegistry()


def get_metrics_registry() -> MetricsRegistry:
    """Get global metrics registry."""
    return _registry


# Common metrics
def record_grpc_request(service: str, method: str, status: str, duration: float):
    """Record gRPC request metrics."""
    # Request counter
    _registry.counter(
        "grpc_requests_total",
        labels={"service": service, "method": method, "status": status},
    ).inc()

    # Request duration histogram
    _registry.histogram(
        "grpc_request_duration_seconds",
        labels={"service": service, "method": method},
    ).observe(duration)


def record_service_health(service: str, healthy: bool):
    """Record service health status."""
    _registry.gauge("service_health", labels={"service": service}).set(
        1.0 if healthy else 0.0
    )


def record_circuit_breaker_state(name: str, state: str):
    """Record circuit breaker state."""
    states = {"closed": 0.0, "half_open": 0.5, "open": 1.0}
    _registry.gauge("circuit_breaker_state", labels={"name": name}).set(
        states.get(state, -1.0)
    )


def record_retry_attempt(service: str, attempt: int, success: bool):
    """Record retry attempt."""
    _registry.counter(
        "retry_attempts_total",
        labels={"service": service, "success": str(success).lower()},
    ).inc()

    if attempt > 1:
        _registry.counter("retry_count_total", labels={"service": service}).inc(
            attempt - 1
        )


class Timer:
    """Context manager for timing operations."""

    def __init__(self, callback):
        """Initialize timer with callback."""
        self.callback = callback
        self.start_time = None

    def __enter__(self):
        """Start timer."""
        self.start_time = time.time()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Stop timer and record duration."""
        if self.start_time is not None:
            duration = time.time() - self.start_time
            self.callback(duration)


def time_grpc_request(service: str, method: str, status: str = "success"):
    """Create timer for gRPC request."""

    def record_duration(duration: float):
        record_grpc_request(service, method, status, duration)

    return Timer(record_duration)
