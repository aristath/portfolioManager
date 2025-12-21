#!/bin/bash
# Auto-deploy script for Arduino Trader
# Runs via cron every 5 minutes to check for updates and deploy changes
# Handles: Main FastAPI app, Arduino app, and sketch rebuilds

set -euo pipefail

# Configuration
REPO_DIR="/home/arduino/repos/autoTrader"
MAIN_APP_DIR="/home/arduino/arduino-trader"
ARDUINO_APP_DIR="/home/arduino/ArduinoApps/trader-display"
LOG_FILE="/home/arduino/logs/auto-deploy.log"
VENV_DIR="$MAIN_APP_DIR/venv"
SERVICE_NAME="arduino-trader"

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S'): $1" >> "$LOG_FILE"
}

# Error handling
error_exit() {
    log "ERROR: $1"
    exit 1
}

# Ensure log directory exists
mkdir -p "$(dirname "$LOG_FILE")"

# Change to repo directory
cd "$REPO_DIR" || error_exit "Cannot change to repo directory: $REPO_DIR"

# Detect current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ -z "$CURRENT_BRANCH" ]; then
    error_exit "Cannot detect current branch"
fi

log "Checking for updates on branch: $CURRENT_BRANCH"

# Fetch latest changes
if ! git fetch origin 2>>"$LOG_FILE"; then
    error_exit "Failed to fetch from origin"
fi

# Get commit hashes
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse "origin/$CURRENT_BRANCH" 2>/dev/null || git rev-parse "origin/main" 2>/dev/null)

if [ -z "$REMOTE" ]; then
    log "WARNING: Cannot find remote branch, skipping update"
    exit 0
fi

# Check if there are changes
if [ "$LOCAL" = "$REMOTE" ]; then
    log "No changes detected (LOCAL: ${LOCAL:0:8} == REMOTE: ${REMOTE:0:8})"
    exit 0
fi

log "Changes detected: LOCAL ${LOCAL:0:8} -> REMOTE ${REMOTE:0:8}"

# Get list of changed files
CHANGED_FILES=$(git diff --name-only "$LOCAL" "$REMOTE" 2>>"$LOG_FILE" || true)

if [ -z "$CHANGED_FILES" ]; then
    log "No changed files detected, but commits differ. Pulling anyway..."
    git pull origin "$CURRENT_BRANCH" >> "$LOG_FILE" 2>&1 || error_exit "Failed to pull changes"
    log "Update complete (no file changes detected)"
    exit 0
fi

log "Changed files: $(echo "$CHANGED_FILES" | tr '\n' ' ')"

# Categorize changes
MAIN_APP_CHANGED=false
ARDUINO_APP_CHANGED=false
SKETCH_CHANGED=false
REQUIREMENTS_CHANGED=false
DEPLOY_SCRIPT_CHANGED=false

while IFS= read -r file; do
    # Check for main app changes
    if [[ "$file" == app/* ]] || \
       [[ "$file" == static/* ]] || \
       [[ "$file" == *.py ]] || \
       [[ "$file" == requirements.txt ]] || \
       [[ "$file" == deploy/arduino-trader.service ]] || \
       [[ "$file" == scripts/* ]] || \
       [[ "$file" == data/* ]]; then
        MAIN_APP_CHANGED=true
        if [[ "$file" == requirements.txt ]]; then
            REQUIREMENTS_CHANGED=true
        fi
    fi
    
    # Check for arduino-app changes
    if [[ "$file" == arduino-app/* ]]; then
        ARDUINO_APP_CHANGED=true

        # Check for sketch changes
        if [[ "$file" == arduino-app/sketch/* ]]; then
            SKETCH_CHANGED=true
        fi

        # Check for deploy script changes
        if [[ "$file" == arduino-app/deploy/auto-deploy.sh ]]; then
            DEPLOY_SCRIPT_CHANGED=true
        fi
    fi
done <<< "$CHANGED_FILES"

# Self-update: Update this script if it changed
if [ "$DEPLOY_SCRIPT_CHANGED" = true ]; then
    log "Deploy script changed - updating self..."
    cp "$REPO_DIR/arduino-app/deploy/auto-deploy.sh" /home/arduino/bin/auto-deploy.sh
    chmod +x /home/arduino/bin/auto-deploy.sh
    log "Deploy script updated"
fi

# Pull latest changes
log "Pulling latest changes from origin/$CURRENT_BRANCH"
if ! git pull origin "$CURRENT_BRANCH" >> "$LOG_FILE" 2>&1; then
    error_exit "Failed to pull changes"
fi

# Deploy main FastAPI app if needed
if [ "$MAIN_APP_CHANGED" = true ]; then
    log "Deploying main FastAPI app..."

    # Sync files to main app directory using cp (rsync not available)
    log "Syncing files to $MAIN_APP_DIR"

    # Create target directory if it doesn't exist
    mkdir -p "$MAIN_APP_DIR"

    # Copy main app directories/files (excluding venv, .env, arduino-app, .git)
    for item in app static scripts data deploy requirements.txt run.py; do
        if [ -e "$REPO_DIR/$item" ]; then
            cp -r "$REPO_DIR/$item" "$MAIN_APP_DIR/" 2>>"$LOG_FILE" || log "WARNING: Failed to copy $item"
        fi
    done

    # Clean up __pycache__ directories
    find "$MAIN_APP_DIR" -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
    find "$MAIN_APP_DIR" -name "*.pyc" -delete 2>/dev/null || true

    log "Main app files synced"
    
    # Update Python dependencies if requirements.txt changed
    if [ "$REQUIREMENTS_CHANGED" = true ]; then
        log "Updating Python dependencies..."
        if [ -d "$VENV_DIR" ]; then
            if source "$VENV_DIR/bin/activate" 2>>"$LOG_FILE"; then
                if pip install -r "$MAIN_APP_DIR/requirements.txt" >> "$LOG_FILE" 2>&1; then
                    log "Dependencies updated successfully"
                else
                    log "WARNING: Failed to update dependencies, continuing anyway"
                fi
            else
                log "WARNING: Failed to activate virtual environment"
            fi
        else
            log "WARNING: Virtual environment not found at $VENV_DIR"
        fi
    fi
    
    # Restart systemd service
    log "Restarting $SERVICE_NAME service"
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        if ! systemctl restart "$SERVICE_NAME" >> "$LOG_FILE" 2>&1; then
            log "WARNING: Failed to restart $SERVICE_NAME service"
        else
            log "$SERVICE_NAME service restarted successfully"
        fi
    else
        log "WARNING: $SERVICE_NAME service is not active, attempting to start"
        systemctl start "$SERVICE_NAME" >> "$LOG_FILE" 2>&1 || log "WARNING: Failed to start $SERVICE_NAME service"
    fi
fi

# Deploy Arduino app if needed
if [ "$ARDUINO_APP_CHANGED" = true ]; then
    log "Deploying Arduino app..."

    # Sync arduino-app files using cp (rsync not available)
    log "Syncing arduino-app files to $ARDUINO_APP_DIR"

    # Create target directory if it doesn't exist
    mkdir -p "$ARDUINO_APP_DIR"

    # Remove old files (except hidden files like .git if any)
    rm -rf "$ARDUINO_APP_DIR"/* 2>/dev/null || true

    # Copy all arduino-app contents
    cp -r "$REPO_DIR/arduino-app/"* "$ARDUINO_APP_DIR/" 2>>"$LOG_FILE" || error_exit "Failed to copy arduino-app files"

    log "Arduino app files synced"
    
    # Rebuild sketch if sketch files changed
    if [ "$SKETCH_CHANGED" = true ]; then
        log "Sketch files changed - triggering rebuild..."
        
        if command -v arduino-app-cli >/dev/null 2>&1; then
            # Stop the app first to ensure clean rebuild
            log "Stopping app for sketch rebuild"
            arduino-app-cli app stop user:trader-display >> "$LOG_FILE" 2>&1 || log "WARNING: Failed to stop app (may not be running)"
            
            # Wait a moment for app to fully stop
            sleep 2
            
            # The Arduino App framework automatically rebuilds the sketch on restart
            # if sketch files have changed. Restart the app to trigger rebuild.
            log "Restarting app to trigger sketch rebuild"
            if arduino-app-cli app restart user:trader-display >> "$LOG_FILE" 2>&1; then
                log "Arduino app restarted - sketch rebuild triggered"
            else
                # If restart fails, try start instead
                log "Restart failed, attempting to start app"
                if arduino-app-cli app start user:trader-display >> "$LOG_FILE" 2>&1; then
                    log "Arduino app started - sketch rebuild triggered"
                else
                    log "ERROR: Failed to start Arduino app after sketch changes"
                fi
            fi
        else
            log "WARNING: arduino-app-cli not found, cannot rebuild sketch"
        fi
    else
        # Just restart the app if no sketch changes
        log "Restarting Arduino app (no sketch changes)"
        if ! arduino-app-cli app restart user:trader-display >> "$LOG_FILE" 2>&1; then
            log "WARNING: Failed to restart Arduino app"
        else
            log "Arduino app restarted successfully"
        fi
    fi
fi

log "Deployment complete"
exit 0
