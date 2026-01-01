#!/bin/bash
# Generate Python code from protobuf definitions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONTRACTS_DIR="$PROJECT_ROOT/contracts"
PROTOS_DIR="$CONTRACTS_DIR/protos"
OUTPUT_DIR="$CONTRACTS_DIR/contracts"

# Activate virtual environment if it exists
if [ -f "$PROJECT_ROOT/venv/bin/activate" ]; then
    source "$PROJECT_ROOT/venv/bin/activate"
fi

echo "Generating Python code from protobuf definitions..."

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate Python code
python3 -m grpc_tools.protoc \
    --proto_path="$PROTOS_DIR" \
    --python_out="$OUTPUT_DIR" \
    --grpc_python_out="$OUTPUT_DIR" \
    "$PROTOS_DIR/common/common.proto" \
    "$PROTOS_DIR/common/position.proto" \
    "$PROTOS_DIR/common/security.proto" \
    "$PROTOS_DIR/planning.proto" \
    "$PROTOS_DIR/scoring.proto" \
    "$PROTOS_DIR/optimization.proto" \
    "$PROTOS_DIR/portfolio.proto" \
    "$PROTOS_DIR/trading.proto" \
    "$PROTOS_DIR/universe.proto" \
    "$PROTOS_DIR/gateway.proto"

# Create __init__.py
cat > "$OUTPUT_DIR/__init__.py" << 'EOF'
"""
Arduino Trader gRPC Contracts.

Auto-generated from protobuf definitions.
Do not edit manually - run scripts/generate_protos.sh to regenerate.
"""

from contracts.common import common_pb2, common_pb2_grpc
from contracts.common import position_pb2, position_pb2_grpc
from contracts.common import security_pb2, security_pb2_grpc
from contracts import (
    planning_pb2,
    planning_pb2_grpc,
    scoring_pb2,
    scoring_pb2_grpc,
    optimization_pb2,
    optimization_pb2_grpc,
    portfolio_pb2,
    portfolio_pb2_grpc,
    trading_pb2,
    trading_pb2_grpc,
    universe_pb2,
    universe_pb2_grpc,
    gateway_pb2,
    gateway_pb2_grpc,
)

__all__ = [
    'common_pb2',
    'common_pb2_grpc',
    'position_pb2',
    'position_pb2_grpc',
    'security_pb2',
    'security_pb2_grpc',
    'planning_pb2',
    'planning_pb2_grpc',
    'scoring_pb2',
    'scoring_pb2_grpc',
    'optimization_pb2',
    'optimization_pb2_grpc',
    'portfolio_pb2',
    'portfolio_pb2_grpc',
    'trading_pb2',
    'trading_pb2_grpc',
    'universe_pb2',
    'universe_pb2_grpc',
    'gateway_pb2',
    'gateway_pb2_grpc',
]
EOF

# Create common package __init__.py
mkdir -p "$OUTPUT_DIR/common"
touch "$OUTPUT_DIR/common/__init__.py"

# Fix imports in generated files
echo "Fixing imports in generated files..."

# Detect OS for sed compatibility (macOS needs -i '', Linux needs -i)
if [[ "$OSTYPE" == "darwin"* ]]; then
    SED_INPLACE="sed -i ''"
else
    SED_INPLACE="sed -i"
fi

# Fix common imports
for file in "$OUTPUT_DIR"/*_pb2.py "$OUTPUT_DIR"/*_pb2_grpc.py "$OUTPUT_DIR/common"/*_pb2.py "$OUTPUT_DIR/common"/*_pb2_grpc.py; do
    if [ -f "$file" ]; then
        $SED_INPLACE 's/^from common import \(.*\)_pb2/from contracts.common import \1_pb2/g' "$file"
    fi
done

# Fix service imports in _grpc.py files (non-common)
for file in "$OUTPUT_DIR"/*_grpc.py; do
    if [ -f "$file" ]; then
        filename=$(basename "$file" _grpc.py)
        $SED_INPLACE "s/^import ${filename}_pb2 as /from contracts import ${filename}_pb2 as /g" "$file"
    fi
done

# Fix service imports in common _grpc.py files
for file in "$OUTPUT_DIR/common"/*_grpc.py; do
    if [ -f "$file" ]; then
        filename=$(basename "$file" _grpc.py)
        $SED_INPLACE "s/^import ${filename}_pb2 as /from contracts.common import ${filename}_pb2 as /g" "$file"
    fi
done

echo "✓ Protobuf code generation complete!"
echo "  Output: $OUTPUT_DIR"

# Install contracts package in development mode
echo "Installing contracts package..."
cd "$CONTRACTS_DIR"
pip install -e .

echo "✓ All done!"
