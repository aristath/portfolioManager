"""Tests for domain exceptions."""

from app.domain.exceptions import DomainError, ValidationError


class TestDomainExceptions:
    """Test domain exception hierarchy."""

    def test_domain_error_is_base_exception(self):
        """Test that DomainError is the base exception."""
        assert issubclass(ValidationError, DomainError)

    def test_validation_error(self):
        """Test ValidationError."""
        error = ValidationError("Symbol cannot be empty")
        assert "Symbol cannot be empty" in str(error)
