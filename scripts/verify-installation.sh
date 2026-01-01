#!/bin/bash
#
# Arduino Trader - Installation Verification Script
#
# Verifies that the microservices installation is configured correctly.
# Can be run after installation to confirm everything is working.
#
# Usage: ./scripts/verify-installation.sh
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
PASSED=0
FAILED=0
WARNINGS=0

print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

print_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

print_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS++))
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Check if file exists
check_file() {
    local file=$1
    local description=$2

    if [ -f "$file" ]; then
        print_pass "$description exists"
        return 0
    else
        print_fail "$description missing: $file"
        return 1
    fi
}

# Check if directory exists
check_dir() {
    local dir=$1
    local description=$2

    if [ -d "$dir" ]; then
        print_pass "$description exists"
        return 0
    else
        print_fail "$description missing: $dir"
        return 1
    fi
}

# Check configuration value
check_config_value() {
    local file=$1
    local key=$2
    local description=$3

    if [ ! -f "$file" ]; then
        print_fail "Config file not found: $file"
        return 1
    fi

    local value=$(grep "^${key}=" "$file" 2>/dev/null | cut -d'=' -f2)

    if [ -z "$value" ]; then
        print_fail "$description not set in $file"
        return 1
    elif [[ "$value" == *"your_"* ]] || [[ "$value" == *"_here"* ]]; then
        print_warn "$description still has placeholder value"
        return 1
    else
        print_pass "$description configured"
        return 0
    fi
}

# Check if service is running (Docker)
check_service_running() {
    local service=$1

    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${service}$"; then
        print_pass "Docker service '$service' is running"
        return 0
    else
        print_fail "Docker service '$service' is not running"
        return 1
    fi
}

# Check if port is listening
check_port() {
    local port=$1
    local service=$2

    if command -v lsof &> /dev/null; then
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            print_pass "$service port $port is listening"
            return 0
        else
            print_fail "$service port $port is not listening"
            return 1
        fi
    elif command -v netstat &> /dev/null; then
        if netstat -tuln 2>/dev/null | grep -q ":$port "; then
            print_pass "$service port $port is listening"
            return 0
        else
            print_fail "$service port $port is not listening"
            return 1
        fi
    else
        print_warn "Cannot check port $port (lsof/netstat not available)"
        return 0
    fi
}

# Check HTTP endpoint
check_http() {
    local url=$1
    local description=$2

    if command -v curl &> /dev/null; then
        if curl -s -f --connect-timeout 5 "$url" > /dev/null 2>&1; then
            print_pass "$description responds to HTTP"
            return 0
        else
            print_fail "$description not responding at $url"
            return 1
        fi
    else
        print_warn "Cannot check $description (curl not available)"
        return 0
    fi
}

print_header "Arduino Trader Installation Verification"

# Check 1: Installation directory
print_header "Installation Files"
check_dir "/home/arduino/arduino-trader" "Installation directory" || true
check_file "/home/arduino/arduino-trader/.env" "Environment file" || true
check_dir "/home/arduino/arduino-trader/app" "Application directory" || true
check_dir "/home/arduino/arduino-trader/services" "Services directory" || true

# Check 2: Configuration files
print_header "Configuration"
check_file "/home/arduino/arduino-trader/app/config/device.yaml" "Device configuration" || true
check_file "/home/arduino/arduino-trader/app/config/services.yaml" "Services configuration" || true
check_file "/home/arduino/arduino-trader/.env" "Environment configuration" || true

# Check API credentials
if [ -f "/home/arduino/arduino-trader/.env" ]; then
    check_config_value "/home/arduino/arduino-trader/.env" "TRADERNET_API_KEY" "Tradernet API Key" || true
    check_config_value "/home/arduino/arduino-trader/.env" "TRADERNET_API_SECRET" "Tradernet API Secret" || true
fi

# Check 3: Docker
print_header "Docker"
if command -v docker &> /dev/null; then
    print_pass "Docker is installed"

    if docker ps &> /dev/null; then
        print_pass "Docker daemon is running"

        # Check docker-compose
        if docker compose version &> /dev/null; then
            print_pass "Docker Compose V2 is available"
        elif docker-compose version &> /dev/null; then
            print_pass "Docker Compose V1 is available"
        else
            print_fail "Docker Compose not found"
        fi
    else
        print_fail "Docker daemon is not running"
    fi
else
    print_fail "Docker is not installed"
fi

# Check 4: Running services
print_header "Microservices"

# Check which services should be running from device.yaml
if [ -f "/home/arduino/arduino-trader/app/config/device.yaml" ]; then
    if command -v yq &> /dev/null; then
        print_info "Reading service configuration..."
        SERVICES=($(yq '.device.roles[]' "/home/arduino/arduino-trader/app/config/device.yaml" 2>/dev/null))

        if [ ${#SERVICES[@]} -gt 0 ]; then
            print_info "Expected services: ${SERVICES[*]}"

            for service in "${SERVICES[@]}"; do
                check_service_running "$service" || true
            done
        else
            print_warn "No services configured in device.yaml"
        fi
    else
        print_warn "yq not installed - cannot read device.yaml (install with: pip install yq)"
    fi
else
    print_warn "device.yaml not found - cannot determine which services should run"
fi

# Check 5: Port availability
print_header "Service Ports"
check_port 8001 "Universe" || true
check_port 8002 "Portfolio" || true
check_port 8003 "Trading" || true
check_port 8004 "Scoring" || true
check_port 8005 "Optimization" || true
check_port 8006 "Planning" || true
check_port 8007 "Gateway" || true
check_port 8000 "Gateway (exposed)" || true

# Check 6: HTTP endpoints
print_header "HTTP Health Checks"
check_http "http://localhost:8000/health" "Gateway" || true
check_http "http://localhost:8001/health" "Universe" || true
check_http "http://localhost:8002/health" "Portfolio" || true
check_http "http://localhost:8003/health" "Trading" || true
check_http "http://localhost:8004/health" "Scoring" || true
check_http "http://localhost:8005/health" "Optimization" || true
check_http "http://localhost:8006/health" "Planning" || true

# Summary
print_header "Verification Summary"
echo ""
echo -e "  Passed:   ${GREEN}$PASSED${NC}"
echo -e "  Failed:   ${RED}$FAILED${NC}"
echo -e "  Warnings: ${YELLOW}$WARNINGS${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ Installation verification PASSED${NC}"
    echo ""
    echo "Next steps:"
    echo "  • Access dashboard: http://localhost:8000"
    echo "  • View API docs: http://localhost:8000/docs"
    echo "  • Check logs: docker compose logs -f"
    exit 0
else
    echo -e "${RED}✗ Installation verification FAILED${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "  • Check docker logs: docker compose logs"
    echo "  • Restart services: docker compose restart"
    echo "  • See INSTALL.md for detailed troubleshooting"
    exit 1
fi
