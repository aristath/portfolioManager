#!/bin/bash
# Deploy and run migration script on Arduino device

set -e

ARDUINO_HOST="arduino@192.168.1.11"
ARDUINO_PASSWORD="aristath"
PYTHON_DATA_DIR="~/arduino-trader/data"
GO_DATA_DIR="~/app/data"

echo "Deploying migration script to Arduino device..."

# Copy migration script to device
scp scripts/migrate_python_to_go.py ${ARDUINO_HOST}:~/arduino-trader/scripts/

echo "Migration script deployed."
echo ""
echo "To run the migration on the Arduino device:"
echo "  ssh ${ARDUINO_HOST}"
echo "  cd ~/arduino-trader"
echo "  python3 scripts/migrate_python_to_go.py --python-data-dir ${PYTHON_DATA_DIR} --go-data-dir ${GO_DATA_DIR}"
echo ""
echo "Or run a dry-run first:"
echo "  python3 scripts/migrate_python_to_go.py --python-data-dir ${PYTHON_DATA_DIR} --go-data-dir ${GO_DATA_DIR} --dry-run"
