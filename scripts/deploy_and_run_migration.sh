#!/bin/bash
# Deploy scripts and run migration on Arduino device

set -e

ARDUINO_HOST="arduino@192.168.1.11"
ARDUINO_PASSWORD="aristath"
PYTHON_DATA_DIR="~/arduino-trader/data"
GO_DATA_DIR="~/app/data"

echo "Deploying migration scripts to Arduino device..."

# Check if sshpass is available
if command -v sshpass &> /dev/null; then
    # Use sshpass for password authentication
    echo "Using sshpass for authentication..."

    # Deploy scripts
    sshpass -p "${ARDUINO_PASSWORD}" scp \
        scripts/migrate_python_to_go.py \
        scripts/backup_legacy_databases.sh \
        scripts/run_migration.sh \
        ${ARDUINO_HOST}:~/arduino-trader/scripts/

    echo "Scripts deployed successfully!"
    echo ""
    echo "Running migration on Arduino device..."
    echo ""

    # Run the migration remotely
    sshpass -p "${ARDUINO_PASSWORD}" ssh ${ARDUINO_HOST} << 'ENDSSH'
cd ~/arduino-trader
chmod +x scripts/*.sh scripts/*.py
./scripts/run_migration.sh
ENDSSH

else
    echo "sshpass not found. Please install it or run manually:"
    echo ""
    echo "To install sshpass:"
    echo "  macOS: brew install hudochenkov/sshpass/sshpass"
    echo "  Linux: sudo apt-get install sshpass"
    echo ""
    echo "Or run manually:"
    echo "  ssh ${ARDUINO_HOST}"
    echo "  cd ~/arduino-trader"
    echo "  # Copy scripts manually or use scp"
    echo "  ./scripts/run_migration.sh"
fi
