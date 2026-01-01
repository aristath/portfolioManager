#!/bin/bash
# Check status of all microservices

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Checking microservices status...${NC}"
echo ""

# Array of services with ports
services=(
    "planning:50051"
    "scoring:50052"
    "optimization:50053"
    "portfolio:50054"
    "trading:50055"
    "universe:50056"
    "gateway:50057"
)

all_running=true

for service_port in "${services[@]}"; do
    IFS=':' read -r service port <<< "$service_port"

    pid_file="$PROJECT_ROOT/logs/$service.pid"

    printf "%-15s " "$service"

    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")

        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${GREEN}✓ RUNNING${NC} (PID: $pid, Port: $port)"
        else
            echo -e "${RED}✗ STOPPED${NC} (stale PID file)"
            all_running=false
        fi
    else
        echo -e "${RED}✗ NOT STARTED${NC}"
        all_running=false
    fi
done

echo ""

if $all_running; then
    echo -e "${GREEN}All services are running!${NC}"
    exit 0
else
    echo -e "${RED}Some services are not running${NC}"
    exit 1
fi
