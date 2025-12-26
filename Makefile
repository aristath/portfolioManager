.PHONY: format lint lint-fix type-check check all

# Format code with Black and isort
format:
	black app/ scripts/ tests/
	isort app/ scripts/ tests/

# Run linters
lint:
	flake8 app/ scripts/ tests/
	mypy app/ --ignore-missing-imports

# Auto-fix linting issues (where possible)
lint-fix:
	black app/ scripts/ tests/
	isort app/ scripts/ tests/

# Type checking only
type-check:
	mypy app/ --ignore-missing-imports

# Run all checks
check: lint type-check

# Format and check
all: format check

