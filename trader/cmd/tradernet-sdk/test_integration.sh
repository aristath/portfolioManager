#!/bin/bash
# Integration test script for Tradernet SDK microservice
# This script tests the microservice with real API credentials

set -e

echo "=== Tradernet SDK Microservice Integration Test ==="
echo ""

# Check if .env.aristath exists
ENV_FILE="../.env.aristath"
if [ ! -f "$ENV_FILE" ]; then
    echo "Error: .env.aristath file not found at $ENV_FILE"
    echo "Please create it with TRADERNET_API_KEY and TRADERNET_API_SECRET"
    exit 1
fi

# Source credentials
source "$ENV_FILE"

if [ -z "$TRADERNET_API_KEY" ] || [ -z "$TRADERNET_API_SECRET" ]; then
    echo "Error: TRADERNET_API_KEY or TRADERNET_API_SECRET not set in .env.aristath"
    exit 1
fi

echo "✓ Credentials loaded"
echo ""

# Build binary
echo "Building binary..."
go build -o tradernet-sdk ./cmd/tradernet-sdk
if [ $? -ne 0 ]; then
    echo "Error: Build failed"
    exit 1
fi
echo "✓ Binary built successfully"
echo ""

# Start server in background
echo "Starting server on port 9001..."
./tradernet-sdk > /tmp/tradernet-sdk.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "Error: Server failed to start"
    cat /tmp/tradernet-sdk.log
    exit 1
fi

echo "✓ Server started (PID: $SERVER_PID)"
echo ""

# Test health endpoint
echo "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:9001/health)
if echo "$HEALTH_RESPONSE" | grep -q '"status":"ok"'; then
    echo "✓ Health check passed"
else
    echo "✗ Health check failed"
    echo "Response: $HEALTH_RESPONSE"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test user-info endpoint
echo "Testing user-info endpoint with real credentials..."
USER_INFO_RESPONSE=$(curl -s -H "X-Tradernet-API-Key: $TRADERNET_API_KEY" \
    -H "X-Tradernet-API-Secret: $TRADERNET_API_SECRET" \
    http://localhost:9001/user-info)

# Check response
if echo "$USER_INFO_RESPONSE" | grep -q '"success":true'; then
    echo "✓ User info request successful"
    echo ""
    echo "Response:"
    echo "$USER_INFO_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$USER_INFO_RESPONSE"
else
    echo "✗ User info request failed"
    echo "Response: $USER_INFO_RESPONSE"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Cleanup
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true
echo "✓ Server stopped"
echo ""

echo "=== Integration test completed successfully ==="
