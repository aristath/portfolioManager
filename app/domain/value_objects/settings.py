"""Settings value objects for application configuration."""

from dataclasses import dataclass
from typing import Dict, Union


@dataclass(frozen=True)
class Settings:
    """Application settings value object.

    Encapsulates all application settings with validation and type safety.

    Note: Planner settings have been moved to TOML configuration files.
    Only global application settings remain here (security scoring, optimizer, cash management).
    """

    # Security scoring
    target_annual_return: float = 0.11
    min_security_score: float = 0.5
    # Portfolio optimizer
    optimizer_blend: float = 0.5
    optimizer_target_return: float = 0.11
    # Cash management
    min_cash_reserve: float = 500.0

    def _validate_non_negative(self, value: float, field_name: str) -> None:
        """Validate that a value is non-negative."""
        if value < 0:
            raise ValueError(f"{field_name} must be non-negative")

    def _validate_positive(self, value: float, field_name: str) -> None:
        """Validate that a value is positive."""
        if value <= 0:
            raise ValueError(f"{field_name} must be positive")

    def _validate_range(
        self, value: float, field_name: str, min_val: float, max_val: float
    ) -> None:
        """Validate that a value is within a range."""
        if not min_val <= value <= max_val:
            raise ValueError(f"{field_name} must be between {min_val} and {max_val}")

    def __post_init__(self):
        """Validate settings values."""
        self._validate_positive(self.target_annual_return, "target_annual_return")
        self._validate_range(self.min_security_score, "min_security_score", 0, 1)
        self._validate_range(self.optimizer_blend, "optimizer_blend", 0, 1)
        self._validate_positive(self.optimizer_target_return, "optimizer_target_return")
        self._validate_non_negative(self.min_cash_reserve, "min_cash_reserve")

    @classmethod
    def from_dict(cls, data: Dict[str, str]) -> "Settings":
        """Create Settings from dictionary (e.g., from repository).

        Args:
            data: Dictionary with setting keys and string values

        Returns:
            Settings instance with parsed values
        """

        def get_float(key: str, default: float) -> float:
            value = data.get(key)
            if value is None:
                return default
            try:
                return float(value)
            except (ValueError, TypeError):
                return default

        return cls(
            target_annual_return=get_float("target_annual_return", 0.11),
            min_security_score=get_float("min_security_score", 0.5),
            optimizer_blend=get_float("optimizer_blend", 0.5),
            optimizer_target_return=get_float("optimizer_target_return", 0.11),
            min_cash_reserve=get_float("min_cash_reserve", 500.0),
        )

    def to_dict(self) -> Dict[str, Union[float, int]]:
        """Convert Settings to dictionary.

        Returns:
            Dictionary with setting keys and typed values
        """
        return {
            "target_annual_return": self.target_annual_return,
            "min_security_score": self.min_security_score,
            "optimizer_blend": self.optimizer_blend,
            "optimizer_target_return": self.optimizer_target_return,
            "min_cash_reserve": self.min_cash_reserve,
        }
