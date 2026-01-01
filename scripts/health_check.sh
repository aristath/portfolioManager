#!/bin/bash
# Health check for all microservices (HTTP/REST)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Services and their HTTP ports
declare -A SERVICES=(
    ["universe"]="8001"
    ["portfolio"]="8002"
    ["trading"]="8003"
    ["scoring"]="8004"
    ["optimization"]="8005"
    ["planning"]="8006"
    ["gateway"]="8007"
    ["opportunity"]="8008"
    ["generator"]="8009"
    ["evaluator-1"]="8010"
    ["coordinator"]="8011"
    ["evaluator-2"]="8020"
    ["evaluator-3"]="8030"
)

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl not installed${NC}"
    echo "Install with: apt-get install curl (Linux) or brew install curl (macOS)"
    exit 1
fi

echo -e "${BLUE}Checking health of all microservices...${NC}"
echo "======================================"
echo ""

HEALTHY_COUNT=0
TOTAL_COUNT=0

# Sort services by port for consistent output
for service in universe portfolio trading scoring optimization planning gateway opportunity generator evaluator-1 coordinator evaluator-2 evaluator-3; do
    port="${SERVICES[$service]}"
    TOTAL_COUNT=$((TOTAL_COUNT + 1))

    # Check if port is listening
    if ! nc -z localhost "$port" 2>/dev/null; then
        echo -e "${RED}✗${NC} $service (HTTP $port): NOT RUNNING"
        continue
    fi

    # Try HTTP health endpoint
    if curl -s -f --connect-timeout 2 "http://localhost:$port/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $service (HTTP $port): HEALTHY"
        HEALTHY_COUNT=$((HEALTHY_COUNT + 1))
    else
        # Port is listening but health endpoint failed
        echo -e "${YELLOW}⚠${NC} $service (HTTP $port): LISTENING (health endpoint unavailable)"
    fi
done

echo ""
echo "======================================"
echo "Health check complete: $HEALTHY_COUNT/$TOTAL_COUNT services healthy"
echo ""

if [ $HEALTHY_COUNT -eq $TOTAL_COUNT ]; then
    echo -e "${GREEN}✓ All services are healthy!${NC}"
    echo ""
    echo "Access points:"
    echo "  • Dashboard: http://localhost:8000"
    echo "  • API Docs:  http://localhost:8000/docs"
    exit 0
elif [ $HEALTHY_COUNT -gt 0 ]; then
    echo -e "${YELLOW}⚠ Some services are not healthy${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "  • Check logs: docker compose logs"
    echo "  • Restart:    docker compose restart"
    exit 1
else
    echo -e "${RED}✗ No services are running!${NC}"
    echo ""
    echo "Start services:"
    echo "  cd /home/arduino/arduino-trader"
    echo "  docker compose up -d"
    exit 2
fi
