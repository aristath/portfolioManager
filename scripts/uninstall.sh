#!/bin/bash
#
# Arduino Trader - Uninstall Script
#
# Safely removes Arduino Trader microservices installation.
# Optionally preserves data and configuration.
#
# Usage:
#   sudo ./scripts/uninstall.sh            # Remove all (prompt for data)
#   sudo ./scripts/uninstall.sh --keep-data # Keep data and configs
#   sudo ./scripts/uninstall.sh --full     # Remove everything
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Default options
KEEP_DATA=false
FULL_UNINSTALL=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --keep-data)
            KEEP_DATA=true
            ;;
        --full)
            FULL_UNINSTALL=true
            ;;
        --help|-h)
            cat << 'EOF'
Arduino Trader - Uninstall Script

USAGE:
    sudo ./scripts/uninstall.sh [OPTIONS]

OPTIONS:
    --keep-data  Preserve data directory and configuration files
    --full       Remove everything including data
    --help, -h   Show this help message

DESCRIPTION:
    Safely uninstall Arduino Trader microservices.
    By default, prompts whether to keep data.

WHAT GETS REMOVED:
    • Docker containers and images
    • Application files (/home/arduino/arduino-trader)
    • Systemd service (if exists)

WHAT CAN BE PRESERVED:
    • Database (data/trader.db)
    • Configuration files (.env, device.yaml, services.yaml)
    • Historical data

EXAMPLES:
    # Interactive (prompts for data preservation)
    sudo ./scripts/uninstall.sh

    # Keep data and configs
    sudo ./scripts/uninstall.sh --keep-data

    # Remove everything
    sudo ./scripts/uninstall.sh --full

EOF
            exit 0
            ;;
    esac
done

# Check root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (sudo ./scripts/uninstall.sh)${NC}"
    exit 1
fi

echo "┌─────────────────────────────────────────┐"
echo "│  Arduino Trader - Uninstall             │"
echo "└─────────────────────────────────────────┘"
echo ""

# Confirm uninstall
echo -e "${YELLOW}This will uninstall Arduino Trader microservices.${NC}"
echo ""
read -p "Are you sure you want to continue? [y/N]: " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled."
    exit 0
fi

# Ask about data if not specified
if [ "$FULL_UNINSTALL" = false ] && [ "$KEEP_DATA" = false ]; then
    echo ""
    echo "Data preservation options:"
    echo "  1. Keep data and configuration (recommended)"
    echo "  2. Remove everything (fresh start)"
    echo ""
    read -p "Keep data and configuration? [Y/n]: " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        KEEP_DATA=true
    fi
fi

APP_DIR="/home/arduino/arduino-trader"
DATA_BACKUP_DIR="/tmp/arduino-trader-backup-$(date +%Y%m%d_%H%M%S)"

# Backup data if requested
if [ "$KEEP_DATA" = true ] && [ -d "$APP_DIR" ]; then
    echo -e "${BLUE}Backing up data and configuration...${NC}"
    mkdir -p "$DATA_BACKUP_DIR"

    # Backup data
    if [ -d "$APP_DIR/data" ]; then
        cp -r "$APP_DIR/data" "$DATA_BACKUP_DIR/"
        echo -e "${GREEN}✓ Backed up data${NC}"
    fi

    # Backup configs
    if [ -f "$APP_DIR/.env" ]; then
        cp "$APP_DIR/.env" "$DATA_BACKUP_DIR/"
        echo -e "${GREEN}✓ Backed up .env${NC}"
    fi

    if [ -d "$APP_DIR/app/config" ]; then
        cp -r "$APP_DIR/app/config" "$DATA_BACKUP_DIR/"
        echo -e "${GREEN}✓ Backed up configuration${NC}"
    fi

    echo -e "${GREEN}✓ Backup saved to: $DATA_BACKUP_DIR${NC}"
fi

# Stop and remove Docker containers
echo ""
echo -e "${BLUE}Stopping Docker containers...${NC}"
if [ -d "$APP_DIR" ]; then
    cd "$APP_DIR"

    if docker compose version &> /dev/null; then
        docker compose down 2>/dev/null || docker-compose down 2>/dev/null || echo "No containers to stop"
    elif docker-compose version &> /dev/null; then
        docker-compose down 2>/dev/null || echo "No containers to stop"
    fi

    echo -e "${GREEN}✓ Containers stopped${NC}"
fi

# Remove Docker images (optional)
read -p "Remove Docker images? [y/N]: " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Removing Docker images...${NC}"
    docker images | grep "arduino-trader" | awk '{print $3}' | xargs -r docker rmi -f 2>/dev/null || true
    echo -e "${GREEN}✓ Images removed${NC}"
fi

# Remove systemd service
echo ""
echo -e "${BLUE}Removing systemd service...${NC}"
if [ -f "/etc/systemd/system/arduino-trader.service" ]; then
    systemctl stop arduino-trader 2>/dev/null || true
    systemctl disable arduino-trader 2>/dev/null || true
    rm -f /etc/systemd/system/arduino-trader.service
    systemctl daemon-reload
    echo -e "${GREEN}✓ Systemd service removed${NC}"
else
    echo "No systemd service found"
fi

# Remove application directory
echo ""
if [ "$KEEP_DATA" = true ]; then
    echo -e "${BLUE}Removing application files (preserving data)...${NC}"
    # Remove everything except data and configs
    if [ -d "$APP_DIR" ]; then
        # Remove app directories
        rm -rf "$APP_DIR/app" 2>/dev/null || true
        rm -rf "$APP_DIR/services" 2>/dev/null || true
        rm -rf "$APP_DIR/static" 2>/dev/null || true
        rm -rf "$APP_DIR/scripts" 2>/dev/null || true
        rm -rf "$APP_DIR/deploy" 2>/dev/null || true
        rm -rf "$APP_DIR/venv" 2>/dev/null || true
        rm -f "$APP_DIR/docker-compose.yml" 2>/dev/null || true
        rm -f "$APP_DIR/requirements.txt" 2>/dev/null || true
        echo -e "${GREEN}✓ Application files removed${NC}"
        echo -e "${YELLOW}ℹ Data preserved in: $APP_DIR/data${NC}"
        echo -e "${YELLOW}ℹ Backup also saved to: $DATA_BACKUP_DIR${NC}"
    fi
else
    echo -e "${BLUE}Removing application directory...${NC}"
    if [ -d "$APP_DIR" ]; then
        rm -rf "$APP_DIR"
        echo -e "${GREEN}✓ Application directory removed${NC}"
    fi
fi

# Summary
echo ""
echo -e "${GREEN}═══════════════════════════════════════════${NC}"
echo -e "${GREEN}  Uninstall Complete!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════${NC}"
echo ""

if [ "$KEEP_DATA" = true ]; then
    echo "Data preserved:"
    echo "  • Backup location: $DATA_BACKUP_DIR"
    echo "  • Original data: $APP_DIR/data (if exists)"
    echo ""
    echo "To reinstall with your data:"
    echo "  1. Run: sudo ./install.sh"
    echo "  2. Copy back: cp -r $DATA_BACKUP_DIR/data $APP_DIR/"
    echo "  3. Copy back: cp $DATA_BACKUP_DIR/.env $APP_DIR/"
else
    echo "All files removed."
    echo ""
    echo "To reinstall:"
    echo "  sudo ./install.sh"
fi

echo ""
