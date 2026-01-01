#!/bin/bash
# Test all REST services startup

cd /Users/aristath/arduino-trader

echo "========================================================================"
echo "Testing All REST Services"
echo "========================================================================"
echo ""

# Test Universe (8001)
echo "Testing universe service (port 8001)..."
cd services/universe
../../venv/bin/python main.py > /tmp/universe.log 2>&1 &
PID=$!
sleep 3
if ps -p $PID > /dev/null 2>&1; then
    response=$(curl -s http://localhost:8001/universe/health 2>/dev/null)
    if echo "$response" | grep -q "healthy"; then
        echo "✅ universe: Service started and responding"
    else
        echo "⚠️  universe: Started but health check failed"
    fi
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
else
    echo "❌ universe: Failed to start"
fi
echo ""

# Test Portfolio (8002)
cd /Users/aristath/arduino-trader
echo "Testing portfolio service (port 8002)..."
cd services/portfolio
../../venv/bin/python main.py > /tmp/portfolio.log 2>&1 &
PID=$!
sleep 3
if ps -p $PID > /dev/null 2>&1; then
    response=$(curl -s http://localhost:8002/portfolio/health 2>/dev/null)
    if echo "$response" | grep -q "healthy"; then
        echo "✅ portfolio: Service started and responding"
    else
        echo "⚠️  portfolio: Started but health check failed"
    fi
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
else
    echo "❌ portfolio: Failed to start"
fi
echo ""

# Test Trading (8003)
cd /Users/aristath/arduino-trader
echo "Testing trading service (port 8003)..."
cd services/trading
../../venv/bin/python main.py > /tmp/trading.log 2>&1 &
PID=$!
sleep 3
if ps -p $PID > /dev/null 2>&1; then
    response=$(curl -s http://localhost:8003/trading/health 2>/dev/null)
    if echo "$response" | grep -q "healthy"; then
        echo "✅ trading: Service started and responding"
    else
        echo "⚠️  trading: Started but health check failed"
    fi
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
else
    echo "❌ trading: Failed to start"
fi
echo ""

# Test Scoring (8004)
cd /Users/aristath/arduino-trader
echo "Testing scoring service (port 8004)..."
cd services/scoring
../../venv/bin/python main.py > /tmp/scoring.log 2>&1 &
PID=$!
sleep 3
if ps -p $PID > /dev/null 2>&1; then
    response=$(curl -s http://localhost:8004/scoring/health 2>/dev/null)
    if echo "$response" | grep -q "healthy"; then
        echo "✅ scoring: Service started and responding"
    else
        echo "⚠️  scoring: Started but health check failed"
    fi
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
else
    echo "❌ scoring: Failed to start"
fi
echo ""

# Test Optimization (8005)
cd /Users/aristath/arduino-trader
echo "Testing optimization service (port 8005)..."
cd services/optimization
../../venv/bin/python main.py > /tmp/optimization.log 2>&1 &
PID=$!
sleep 3
if ps -p $PID > /dev/null 2>&1; then
    response=$(curl -s http://localhost:8005/optimization/health 2>/dev/null)
    if echo "$response" | grep -q "healthy"; then
        echo "✅ optimization: Service started and responding"
    else
        echo "⚠️  optimization: Started but health check failed"
    fi
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
else
    echo "❌ optimization: Failed to start"
fi
echo ""

# Test Planning (8006)
cd /Users/aristath/arduino-trader
echo "Testing planning service (port 8006)..."
cd services/planning
../../venv/bin/python main.py > /tmp/planning.log 2>&1 &
PID=$!
sleep 3
if ps -p $PID > /dev/null 2>&1; then
    response=$(curl -s http://localhost:8006/planning/health 2>/dev/null)
    if echo "$response" | grep -q "healthy"; then
        echo "✅ planning: Service started and responding"
    else
        echo "⚠️  planning: Started but health check failed"
    fi
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
else
    echo "❌ planning: Failed to start"
fi
echo ""

# Test Gateway (8007)
cd /Users/aristath/arduino-trader
echo "Testing gateway service (port 8007)..."
cd services/gateway
../../venv/bin/python main.py > /tmp/gateway.log 2>&1 &
PID=$!
sleep 3
if ps -p $PID > /dev/null 2>&1; then
    response=$(curl -s http://localhost:8007/gateway/health 2>/dev/null)
    if echo "$response" | grep -q "healthy"; then
        echo "✅ gateway: Service started and responding"
    else
        echo "⚠️  gateway: Started but health check failed"
    fi
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
else
    echo "❌ gateway: Failed to start"
fi
echo ""

echo "========================================================================"
echo "Test Complete"
echo "========================================================================"
