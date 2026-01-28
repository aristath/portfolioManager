#!/bin/bash
# Generate version string from current UTC date+time
# Format: vYYYY.MM.DD.HH.MM

VERSION="v$(date -u +%Y.%m.%d.%H.%M)"

cat > sentinel/version.py << EOF
VERSION = "$VERSION"
EOF
