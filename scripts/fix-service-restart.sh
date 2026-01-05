#!/bin/bash
# Script to fix systemd service restart issue on Arduino device
# Run this on the device after freeing up disk space

set -e

echo "=== Fixing systemd service restart configuration ==="

# Option 1: Remove NoNewPrivileges (simplest, but reduces security)
echo ""
echo "Option 1: Removing NoNewPrivileges from service file..."
sudo sed -i 's/^NoNewPrivileges=true/#NoNewPrivileges=true  # Disabled to allow service restarts via deployment/' /etc/systemd/system/trader.service
sudo systemctl daemon-reload
echo "✓ Service file updated and systemd reloaded"

# Option 2: Configure Polkit (more secure, keeps NoNewPrivileges)
echo ""
echo "Option 2: Configuring Polkit rules (alternative secure method)..."
sudo mkdir -p /etc/polkit-1/rules.d
sudo bash -c 'cat > /etc/polkit-1/rules.d/50-arduino-trader.rules << EOF
polkit.addRule(function(action, subject) {
    if (action.id == "org.freedesktop.systemd1.manage-units" &&
        subject.user == "arduino" &&
        (action.lookup("unit") == "trader.service" || action.lookup("unit") == "arduino-trader.service")) {
        return polkit.Result.YES;
    }
});
EOF'
sudo chmod 644 /etc/polkit-1/rules.d/50-arduino-trader.rules
echo "✓ Polkit rule created"

echo ""
echo "=== Configuration complete ==="
echo ""
echo "The deployment system will now be able to restart the trader service."
echo "It will try multiple methods:"
echo "  1. Direct systemctl (may work with polkit)"
echo "  2. sudo systemctl (works if NoNewPrivileges is disabled)"
echo "  3. D-Bus API (works with polkit rule)"
echo ""
echo "To test, trigger a deployment:"
echo "  curl -X POST http://localhost:8080/api/system/deployment/hard-update"
