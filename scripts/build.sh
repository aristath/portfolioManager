#!/bin/bash
# Build Go applications for Arduino Uno Q (ARM64 Linux)

set -e  # Exit on error

# Load configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.sh"

cd "$(dirname "$SCRIPT_DIR")"  # Change to repo root

log_header "Building for Arduino Uno Q (ARM64)"

# Build trader-go
if [ "$BUILD_TRADER_GO" = "yes" ]; then
    log_info "Building trader-go..."
    cd trader-go/cmd/server
    GOOS=linux GOARCH=arm64 go build -o ../../../build/trader-go
    cd ../../..
    log_success "trader-go built successfully → build/trader-go"
fi

# Build bridge-go
if [ "$BUILD_BRIDGE_GO" = "yes" ]; then
    log_info "Building bridge-go..."
    cd bridge-go
    GOOS=linux GOARCH=arm64 go build -o ../build/bridge-go
    cd ..
    log_success "bridge-go built successfully → build/bridge-go"
fi

log_header "Build Summary"
ls -lh build/
log_success "All builds complete!"
