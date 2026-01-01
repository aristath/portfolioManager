#!/bin/bash
# Start all microservices locally

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting all microservices...${NC}"

# Check if virtual environment exists
if [ ! -d "$PROJECT_ROOT/venv" ]; then
    echo -e "${YELLOW}Virtual environment not found. Creating...${NC}"
    python3 -m venv "$PROJECT_ROOT/venv"
    source "$PROJECT_ROOT/venv/bin/activate"
    pip install -r "$PROJECT_ROOT/requirements.txt"
else
    source "$PROJECT_ROOT/venv/bin/activate"
fi

# Set Python path
export PYTHONPATH="$PROJECT_ROOT:$PYTHONPATH"

# Set config paths
export DEVICE_CONFIG_PATH="$PROJECT_ROOT/app/config/device.yaml"
export SERVICES_CONFIG_PATH="$PROJECT_ROOT/app/config/services.yaml"

# Create logs directory
mkdir -p "$PROJECT_ROOT/logs"

# Array of gRPC services (with their gRPC ports)
# Note: The new Planning microservices (opportunity, generator, coordinator, evaluator-1/2/3)
# are HTTP-only (FastAPI) and don't use gRPC, so they are not included here.
# They run on HTTP ports 8008-8011, 8020, 8030 and are managed via Docker Compose.
services=(
    "planning:50051"
    "scoring:50052"
    "optimization:50053"
    "portfolio:50054"
    "trading:50055"
    "universe:50056"
    "gateway:50057"
)

echo -e "${GREEN}Starting services...${NC}"

for service_port in "${services[@]}"; do
    IFS=':' read -r service port <<< "$service_port"

    echo -e "${YELLOW}Starting $service service on port $port...${NC}"

    nohup python3 -m "services.$service.main" \
        > "$PROJECT_ROOT/logs/$service.log" 2>&1 &

    echo $! > "$PROJECT_ROOT/logs/$service.pid"

    # Wait a bit for service to start
    sleep 1

    # Check if service is still running
    if kill -0 $(cat "$PROJECT_ROOT/logs/$service.pid") 2>/dev/null; then
        echo -e "${GREEN}✓ $service service started (PID: $(cat "$PROJECT_ROOT/logs/$service.pid"))${NC}"
    else
        echo -e "${RED}✗ $service service failed to start${NC}"
        cat "$PROJECT_ROOT/logs/$service.log"
    fi
done

echo -e "${GREEN}All services started!${NC}"
echo -e "${YELLOW}Check logs in: $PROJECT_ROOT/logs/${NC}"
echo -e "${YELLOW}Use deploy/scripts/stop-all-services.sh to stop services${NC}"
