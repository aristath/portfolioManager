#!/bin/bash
# Stop all microservices

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Stopping all microservices...${NC}"

# Array of services
services=(
    "planning"
    "scoring"
    "optimization"
    "portfolio"
    "trading"
    "universe"
    "gateway"
)

for service in "${services[@]}"; do
    pid_file="$PROJECT_ROOT/logs/$service.pid"

    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")

        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${YELLOW}Stopping $service service (PID: $pid)...${NC}"
            kill "$pid"

            # Wait for graceful shutdown
            timeout=5
            while kill -0 "$pid" 2>/dev/null && [ $timeout -gt 0 ]; do
                sleep 1
                timeout=$((timeout - 1))
            done

            if kill -0 "$pid" 2>/dev/null; then
                echo -e "${RED}Force killing $service service...${NC}"
                kill -9 "$pid"
            fi

            echo -e "${GREEN}✓ $service service stopped${NC}"
        else
            echo -e "${YELLOW}$service service not running (stale PID file)${NC}"
        fi

        rm "$pid_file"
    else
        echo -e "${YELLOW}$service service PID file not found${NC}"
    fi
done

echo -e "${GREEN}All services stopped!${NC}"
