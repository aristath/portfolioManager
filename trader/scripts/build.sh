#!/bin/bash
set -e

# Build script for Arduino Trader Go

ARCH=${1:-amd64}  # Default to amd64, can specify arm64
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

echo "Building Arduino Trader Go..."
echo "Architecture: ${ARCH}"
echo "Version: ${VERSION}"
echo "Build Time: ${BUILD_TIME}"
echo

echo "Building for local architecture..."
go build -ldflags "${LDFLAGS}" -o trader ./cmd/server
echo "✓ Built: trader-go"

echo
echo "Build complete!"
