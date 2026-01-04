#!/bin/bash
# Complete migration workflow: backup legacy databases, then migrate to Go app

set -e

PYTHON_DATA_DIR="${1:-~/arduino-trader/data}"
GO_DATA_DIR="${2:-~/app/data}"
DRY_RUN="${3:-}"

echo "=========================================="
echo "Python to Go Database Migration"
echo "=========================================="
echo "Python data dir: ${PYTHON_DATA_DIR}"
echo "Go data dir: ${GO_DATA_DIR}"
echo ""

# Step 1: Backup legacy databases
echo "Step 1: Backing up legacy Python databases..."
./scripts/backup_legacy_databases.sh "${PYTHON_DATA_DIR}"

echo ""
echo "Step 2: Running migration..."

# Step 2: Run migration
if [ -n "${DRY_RUN}" ]; then
    echo "  Running in DRY-RUN mode..."
    python3 scripts/migrate_python_to_go.py \
        --python-data-dir "${PYTHON_DATA_DIR}" \
        --go-data-dir "${GO_DATA_DIR}" \
        --dry-run
else
    echo "  Running actual migration..."
    python3 scripts/migrate_python_to_go.py \
        --python-data-dir "${PYTHON_DATA_DIR}" \
        --go-data-dir "${GO_DATA_DIR}"
fi

echo ""
echo "=========================================="
echo "Migration completed!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Verify the Go app can read the migrated data"
echo "2. Test critical functions (portfolio, trading, planning)"
echo "3. Monitor for any issues"
echo "4. Once confirmed working, you can stop the Python app"
