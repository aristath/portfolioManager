"""Domain-specific exceptions."""


class DomainError(Exception):
    """Base exception for domain errors."""


class ValidationError(DomainError):
    """Raised when domain validation fails."""

    def __init__(self, message: str):
        self.message = message
        super().__init__(f"Validation error: {message}")
