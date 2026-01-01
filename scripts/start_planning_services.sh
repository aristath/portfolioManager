#!/bin/bash
# Start Planning microservices
# Usage: ./scripts/start_planning_services.sh [build|up|down|logs|health]

set -e

SERVICE_NAMES="opportunity generator evaluator-1 evaluator-2 evaluator-3 coordinator"

function build_services() {
    echo "Building Planning microservices..."
    docker-compose build $SERVICE_NAMES
}

function start_services() {
    echo "Starting Planning microservices..."
    docker-compose up -d $SERVICE_NAMES
    echo "Waiting for services to be healthy..."
    sleep 5
    check_health
}

function stop_services() {
    echo "Stopping Planning microservices..."
    docker-compose down $SERVICE_NAMES
}

function show_logs() {
    docker-compose logs -f $SERVICE_NAMES
}

function check_health() {
    echo "Checking service health..."

    services=(
        "http://localhost:8008/opportunity/health:Opportunity"
        "http://localhost:8009/generator/health:Generator"
        "http://localhost:8010/evaluator/health:Evaluator-1"
        "http://localhost:8020/evaluator/health:Evaluator-2"
        "http://localhost:8030/evaluator/health:Evaluator-3"
        "http://localhost:8011/coordinator/health:Coordinator"
    )

    all_healthy=true
    for service in "${services[@]}"; do
        IFS=':' read -r url name <<< "$service"
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo "✓ $name is healthy"
        else
            echo "✗ $name is not responding"
            all_healthy=false
        fi
    done

    if [ "$all_healthy" = true ]; then
        echo ""
        echo "✅ All services are healthy!"
        echo ""
        echo "Next steps:"
        echo "  - Run integration tests: ./venv/bin/python3 -m pytest tests/integration/services/planning/ -v"
        echo "  - Run equivalence test: ./venv/bin/python3 -m pytest tests/integration/services/planning/test_planning_equivalence.py -v"
        echo "  - Run performance test: ./venv/bin/python3 -m pytest tests/performance/test_planning_microservices_performance.py -v"
    else
        echo ""
        echo "⚠️ Some services are not healthy. Check logs with: $0 logs"
        exit 1
    fi
}

case "${1:-help}" in
    build)
        build_services
        ;;
    up|start)
        start_services
        ;;
    down|stop)
        stop_services
        ;;
    logs)
        show_logs
        ;;
    health)
        check_health
        ;;
    restart)
        stop_services
        sleep 2
        start_services
        ;;
    help|*)
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  build     - Build all Planning microservices"
        echo "  up|start  - Start all Planning microservices"
        echo "  down|stop - Stop all Planning microservices"
        echo "  logs      - Show logs from all services"
        echo "  health    - Check health of all services"
        echo "  restart   - Restart all services"
        echo ""
        echo "Examples:"
        echo "  $0 build        # Build services"
        echo "  $0 up           # Start services"
        echo "  $0 health       # Check if services are healthy"
        echo "  $0 logs         # Follow logs"
        ;;
esac
