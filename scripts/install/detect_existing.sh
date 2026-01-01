#!/bin/bash
#
# Detect Existing Installation
# Checks for existing Arduino Trader installation and reads current configuration
#

detect_existing_installation() {
    local app_dir="/home/arduino/arduino-trader"
    local env_file="${app_dir}/.env"
    local device_config="${app_dir}/app/config/device.yaml"
    local services_config="${app_dir}/app/config/services.yaml"

    # Check if this is an existing installation
    if [ -f "$env_file" ] && [ -f "$device_config" ] && [ -f "$services_config" ]; then
        INSTALL_TYPE="existing"

        echo "  → Found existing installation!"

        # Read current device roles
        if command -v yq &> /dev/null; then
            local roles=$(yq '.device.roles[]' "$device_config" 2>/dev/null | tr '\n' ' ')
            if [ -n "$roles" ]; then
                echo "  → Current services on this device: $roles"
            fi
        else
            echo "  → Current installation detected (install yq to see service details)"
        fi

        # Determine deployment mode
        local mode="Unknown"
        if grep -q 'mode: "local"' "$services_config" 2>/dev/null; then
            mode="Single-device"
        elif grep -q 'mode: "remote"' "$services_config" 2>/dev/null; then
            mode="Distributed"
        fi
        echo "  → Device mode: $mode"

        # Ask if user wants to modify
        echo ""
        read -p "Do you want to modify this installation? [y/N]: " -n 1 -r
        echo ""

        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_msg "${YELLOW}" "Installation cancelled."
            exit 0
        fi
    else
        INSTALL_TYPE="fresh"
        echo "  → No existing installation found"
        echo "  → Starting fresh installation"
    fi
}
