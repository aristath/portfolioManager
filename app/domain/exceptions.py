"""Domain-specific exceptions."""


class DomainError(Exception):
    """Base exception for domain errors."""
    pass


class StockNotFoundError(DomainError):
    """Raised when a stock is not found."""
    pass


class InsufficientCashError(DomainError):
    """Raised when there's insufficient cash for a trade."""
    pass


class TradeExecutionError(DomainError):
    """Raised when trade execution fails."""
    pass


class InvalidAllocationError(DomainError):
    """Raised when allocation target is invalid."""
    pass


class ScoreCalculationError(DomainError):
    """Raised when stock score calculation fails."""
    pass
