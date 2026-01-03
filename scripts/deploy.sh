#!/bin/bash
# Deploy binaries to Arduino Uno Q device

set -e  # Exit on error

# Load configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.sh"

cd "$(dirname "$SCRIPT_DIR")"  # Change to repo root

log_header "Deploying to Arduino Uno Q"
log_info "Target: ${ARDUINO_SSH}:${ARDUINO_DEPLOY_PATH}"

# Ensure build directory exists
if [ ! -d "build" ]; then
    log_error "Build directory not found. Run ./scripts/build.sh first."
    exit 1
fi

# Check connectivity
log_info "Testing SSH connection..."
if ! ssh -o ConnectTimeout=5 "${ARDUINO_SSH}" "echo 'Connection successful'" >/dev/null 2>&1; then
    log_error "Cannot connect to ${ARDUINO_SSH}"
    log_info "Ensure the Arduino device is powered on and accessible"
    exit 1
fi
log_success "SSH connection verified"

# Create deployment directory if it doesn't exist
log_info "Ensuring deployment directory exists..."
ssh "${ARDUINO_SSH}" "mkdir -p ${ARDUINO_DEPLOY_PATH}"

# Deploy trader-go
if [ -f "build/trader-go" ]; then
    log_info "Deploying trader-go..."
    scp build/trader-go "${ARDUINO_SSH}:${ARDUINO_DEPLOY_PATH}/"
    ssh "${ARDUINO_SSH}" "chmod +x ${ARDUINO_DEPLOY_PATH}/trader-go"
    log_success "trader-go deployed"
else
    log_warn "build/trader-go not found, skipping"
fi

# Deploy bridge-go
if [ -f "build/bridge-go" ]; then
    log_info "Deploying bridge-go..."
    scp build/bridge-go "${ARDUINO_SSH}:${ARDUINO_DEPLOY_PATH}/"
    ssh "${ARDUINO_SSH}" "chmod +x ${ARDUINO_DEPLOY_PATH}/bridge-go"
    log_success "bridge-go deployed"
else
    log_warn "build/bridge-go not found, skipping"
fi

# Deploy data directory (if needed)
if [ "$DEPLOY_DATA" = "yes" ]; then
    log_info "Deploying data directory..."
    scp -r data/ "${ARDUINO_SSH}:${ARDUINO_DEPLOY_PATH}/"
    log_success "Data directory deployed"
fi

log_header "Deployment Summary"
ssh "${ARDUINO_SSH}" "ls -lh ${ARDUINO_DEPLOY_PATH}/ | grep -E 'trader-go|bridge-go|data'"
log_success "Deployment complete!"
log_info "Run './scripts/restart.sh' to restart services"
