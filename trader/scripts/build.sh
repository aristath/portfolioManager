#!/bin/bash
set -e

# Build script for Sentinel Go

ARCH=${1:-amd64}  # Default to amd64, can specify arm64
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

echo "Building Sentinel Go..."
echo "Architecture: ${ARCH}"
echo "Version: ${VERSION}"
echo "Build Time: ${BUILD_TIME}"
echo


if [ "$ARCH" = "arm64" ]; then
    echo "Cross-compiling for ARM64 (Arduino Uno Q)..."
    GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o sentinel-arm64 ./cmd/server
    echo "✓ Built: sentinel-arm64"
else
    echo "Building for local architecture..."
    go build -ldflags "${LDFLAGS}" -o sentinel ./cmd/server
    echo "✓ Built: sentinel"
fi

echo
echo "Build complete!"
