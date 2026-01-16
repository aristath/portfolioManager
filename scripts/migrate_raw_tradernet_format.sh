#!/bin/bash
set -e

DB_PATH="$HOME/data/universe.db"
API_HOST="${API_HOST:-localhost:8001}"

echo "Starting migration to raw Tradernet format..."
echo "API Host: $API_HOST"

# Get all ISINs and symbols
SECURITIES=$(sqlite3 "$DB_PATH" "SELECT isin, symbol FROM securities" 2>/dev/null)

if [ -z "$SECURITIES" ]; then
    echo "ERROR: No securities found in database"
    exit 1
fi

TOTAL=$(echo "$SECURITIES" | wc -l)
CURRENT=0

echo "Found $TOTAL securities to migrate"
echo ""

echo "$SECURITIES" | while IFS='|' read -r isin symbol; do
    CURRENT=$((CURRENT + 1))
    echo "[$CURRENT/$TOTAL] Migrating $symbol ($isin)..."

    # Trigger metadata sync for this security via API
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null -X POST "http://$API_HOST/api/work/trigger/security:metadata/$isin" 2>/dev/null || echo "000")

    if [ "$HTTP_CODE" != "200" ] && [ "$HTTP_CODE" != "202" ]; then
        echo "  WARNING: Failed to sync $symbol (HTTP $HTTP_CODE), continuing..."
        continue
    fi

    echo "  ✓ Synced $symbol"

    # Wait 1.5 seconds between requests (rate limiting)
    sleep 1.5
done

echo ""
echo "Migration complete. Verifying..."

# Verify all securities have last_synced set
NOT_SYNCED=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM securities WHERE last_synced IS NULL" 2>/dev/null)
if [ "$NOT_SYNCED" -gt 0 ]; then
    echo "WARNING: $NOT_SYNCED securities not synced"
    echo "You may need to re-run the migration or sync manually via the UI"
    exit 0
fi

echo "✓ All securities migrated successfully"
