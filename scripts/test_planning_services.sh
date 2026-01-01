#!/bin/bash
# Test Planning microservices
# Usage: ./scripts/test_planning_services.sh [unit|integration|equivalence|performance|all]

set -e

cd "$(dirname "$0")/.."

function run_unit_tests() {
    echo "Running unit tests for Planning microservices..."
    ./venv/bin/python3 -m pytest \
        tests/unit/services/opportunity/ \
        tests/unit/services/generator/ \
        tests/unit/services/evaluator/ \
        tests/unit/services/coordinator/ \
        -v --tb=short
}

function run_integration_tests() {
    echo "Running integration tests for Planning microservices..."
    echo "Note: Requires services to be running (docker-compose up)"
    ./venv/bin/python3 -m pytest \
        tests/integration/services/planning/test_planning_microservices.py \
        -v --tb=short
}

function run_equivalence_test() {
    echo "Running equivalence test (microservices vs monolithic)..."
    echo "Note: Requires services to be running (docker-compose up)"
    ./venv/bin/python3 -m pytest \
        tests/integration/services/planning/test_planning_equivalence.py \
        -v --tb=short
}

function run_performance_test() {
    echo "Running performance test (verifying 2.5× speedup)..."
    echo "Note: Requires services to be running (docker-compose up)"
    ./venv/bin/python3 -m pytest \
        tests/performance/test_planning_microservices_performance.py \
        -v --tb=short
}

function run_all_tests() {
    echo "Running all Planning microservices tests..."
    echo ""
    echo "=== Unit Tests ==="
    run_unit_tests
    echo ""
    echo "=== Integration Tests ==="
    run_integration_tests
    echo ""
    echo "=== Equivalence Test ==="
    run_equivalence_test
    echo ""
    echo "=== Performance Test ==="
    run_performance_test
}

case "${1:-help}" in
    unit)
        run_unit_tests
        ;;
    integration)
        run_integration_tests
        ;;
    equivalence)
        run_equivalence_test
        ;;
    performance)
        run_performance_test
        ;;
    all)
        run_all_tests
        ;;
    help|*)
        echo "Usage: $0 [test-type]"
        echo ""
        echo "Test Types:"
        echo "  unit         - Run unit tests (no services required)"
        echo "  integration  - Run integration tests (requires running services)"
        echo "  equivalence  - Run equivalence test (requires running services)"
        echo "  performance  - Run performance test (requires running services)"
        echo "  all          - Run all tests"
        echo ""
        echo "Examples:"
        echo "  $0 unit              # Run unit tests only"
        echo "  $0 integration       # Run integration tests"
        echo "  $0 all               # Run all tests"
        echo ""
        echo "Before running integration/equivalence/performance tests:"
        echo "  ./scripts/start_planning_services.sh up"
        ;;
esac
