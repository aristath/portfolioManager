#!/bin/bash
#
# Arduino Trader - Quick Status Dashboard
#
# Shows a quick overview of system status including:
# - Docker services status
# - Service health
# - Resource usage
# - Recent logs
#
# Usage: ./scripts/status.sh
#

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

clear
echo "┌─────────────────────────────────────────┐"
echo "│  Arduino Trader - Status Dashboard      │"
echo "└─────────────────────────────────────────┘"

# Docker Services
print_header "Docker Services"
if command -v docker &> /dev/null; then
    if docker compose version &> /dev/null 2>&1; then
        cd /home/arduino/arduino-trader 2>/dev/null || cd "$(dirname "$(dirname "$0")")"

        # Get service status
        echo ""
        docker compose ps --format "table {{.Service}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || \
            echo "No services found or docker-compose.yml not in current directory"
    else
        echo -e "${YELLOW}Docker Compose not available${NC}"
    fi
else
    echo -e "${RED}Docker not installed${NC}"
fi

# Service Health
print_header "Service Health"
if command -v curl &> /dev/null; then
    echo ""

    # Check each service
    services=("universe:8001" "portfolio:8002" "trading:8003" "scoring:8004" "optimization:8005" "planning:8006" "gateway:8007" "opportunity:8008" "generator:8009" "evaluator-1:8010" "coordinator:8011" "evaluator-2:8020" "evaluator-3:8030")

    for svc in "${services[@]}"; do
        IFS=':' read -r name port <<< "$svc"

        if curl -s -f --connect-timeout 1 "http://localhost:$port/health" > /dev/null 2>&1; then
            printf "  ${GREEN}✓${NC} %-13s (:%s)\n" "$name" "$port"
        else
            printf "  ${RED}✗${NC} %-13s (:%s)\n" "$name" "$port"
        fi
    done
else
    echo "curl not available - cannot check HTTP endpoints"
fi

# Resource Usage
print_header "Resource Usage"
if command -v docker &> /dev/null; then
    echo ""
    echo "Container Stats (1 second sample):"
    echo ""
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | head -10
else
    echo "Docker not available"
fi

# Disk Usage
if [ -d "/home/arduino/arduino-trader/data" ]; then
    echo ""
    echo "Data Directory:"
    du -sh /home/arduino/arduino-trader/data 2>/dev/null || echo "Unable to check disk usage"
fi

# Recent Activity
print_header "Recent Activity (last 10 logs)"
if command -v docker &> /dev/null; then
    cd /home/arduino/arduino-trader 2>/dev/null || cd "$(dirname "$(dirname "$0")")"
    echo ""
    docker compose logs --tail=10 2>/dev/null | tail -20 || echo "No logs available"
else
    echo "Docker not available"
fi

# Quick Actions
print_header "Quick Actions"
echo ""
echo "  View logs:       docker compose logs -f"
echo "  Restart all:     docker compose restart"
echo "  Stop all:        docker compose stop"
echo "  Start all:       docker compose up -d"
echo "  Health check:    ./scripts/health_check.sh"
echo "  Full verify:     ./scripts/verify-installation.sh"
echo ""
echo "  Dashboard:       http://localhost:8000"
echo "  API Docs:        http://localhost:8000/docs"
echo ""
