#!/bin/bash
# Auto-deploy script for Arduino Trader
# Runs via cron every 5 minutes to check for updates and deploy changes
# Handles: Main FastAPI app and sketch compilation/upload

set -euo pipefail

# Configuration
REPO_DIR="/home/arduino/repos/autoTrader"
MAIN_APP_DIR="/home/arduino/arduino-trader"
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

# Restart systemd service with retry logic
restart_service() {
    local max_attempts=3
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        log "Attempt $attempt/$max_attempts: Restarting $SERVICE_NAME service"

        if sudo systemctl restart "$SERVICE_NAME" 2>>"$LOG_FILE"; then
            # Wait for service to start
            sleep 2

            # Verify service is running
            if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
                log "$SERVICE_NAME service restarted successfully"
                return 0
            else
                log "WARNING: Restart command succeeded but service is not active"
            fi
        else
            log "WARNING: Failed to restart $SERVICE_NAME (attempt $attempt)"
        fi

        attempt=$((attempt + 1))
        if [ $attempt -le $max_attempts ]; then
            log "Waiting 5 seconds before retry..."
            sleep 5
        fi
    done

    log "ERROR: Failed to restart $SERVICE_NAME after $max_attempts attempts"
    return 1
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
SKETCH_CHANGED=false
REQUIREMENTS_CHANGED=false
DEPLOY_SCRIPT_CHANGED=false
MICROSERVICES_CHANGED=()

# Map file patterns to Docker service names
declare -A FILE_TO_SERVICE_MAP=(
    # Planning microservices
    ["services/opportunity/"]="opportunity"
    ["services/generator/"]="generator"
    ["services/coordinator/"]="coordinator"
    ["services/evaluator/"]="evaluator-1 evaluator-2 evaluator-3"

    # Shared Planning code affects all Planning microservices
    ["app/modules/planning/domain/"]="opportunity generator coordinator evaluator-1 evaluator-2 evaluator-3"
    ["app/modules/planning/services/local_opportunity"]="opportunity"
    ["app/modules/planning/services/local_generator"]="generator"
    ["app/modules/planning/services/local_coordinator"]="coordinator"
    ["app/modules/planning/services/local_evaluator"]="evaluator-1 evaluator-2 evaluator-3"

    # Original services (gRPC-based)
    ["services/planning/"]="planning"
    ["services/scoring/"]="scoring"
    ["services/optimization/"]="optimization"
    ["services/portfolio/"]="portfolio"
    ["services/trading/"]="trading"
    ["services/universe/"]="universe"
    ["services/gateway/"]="gateway"
)

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

    # Check for sketch changes
    if [[ "$file" == arduino-app/sketch/* ]]; then
        SKETCH_CHANGED=true
    fi

    # Check for deploy script changes
    if [[ "$file" == arduino-app/deploy/auto-deploy.sh ]]; then
        DEPLOY_SCRIPT_CHANGED=true
    fi

    # Check for microservice changes
    for pattern in "${!FILE_TO_SERVICE_MAP[@]}"; do
        if [[ "$file" == $pattern* ]]; then
            # Add affected services to restart list
            services="${FILE_TO_SERVICE_MAP[$pattern]}"
            for service in $services; do
                # Check if already in list to avoid duplicates
                if [[ ! " ${MICROSERVICES_CHANGED[@]} " =~ " ${service} " ]]; then
                    MICROSERVICES_CHANGED+=("$service")
                fi
            done
        fi
    done
done <<< "$CHANGED_FILES"

# Self-update: Update this script if it changed
if [ "$DEPLOY_SCRIPT_CHANGED" = true ]; then
    log "Deploy script changed - updating self..."
    cp "$REPO_DIR/arduino-app/deploy/auto-deploy.sh" /home/arduino/bin/auto-deploy.sh
    chmod +x /home/arduino/bin/auto-deploy.sh
    log "Deploy script updated"
fi

# Reset any local changes (device should always match remote)
log "Resetting local changes..."
git reset --hard HEAD >> "$LOG_FILE" 2>&1 || log "WARNING: git reset failed"
git clean -fd >> "$LOG_FILE" 2>&1 || log "WARNING: git clean failed"

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

    # Reset deployment directory git repo if it exists (to prevent showing modified files)
    # This ensures the deployment directory's git state matches the repo after file copy
    if [ -d "$MAIN_APP_DIR/.git" ]; then
        log "Resetting deployment directory git repo to match repo state..."
        (
            cd "$MAIN_APP_DIR" || exit 1
            # Fetch latest to ensure we have the same commits
            git fetch origin >> "$LOG_FILE" 2>&1 || log "WARNING: git fetch failed in deployment directory"
            # Reset to match the repo's current commit
            git reset --hard "$(cd "$REPO_DIR" && git rev-parse HEAD)" >> "$LOG_FILE" 2>&1 || log "WARNING: git reset failed in deployment directory"
            git clean -fd >> "$LOG_FILE" 2>&1 || log "WARNING: git clean failed in deployment directory"
        ) || log "WARNING: Failed to reset deployment directory git repo"
    fi

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

    # Restart systemd service with retry logic
    restart_service
fi

# Handle microservice restarts (device-aware, intelligent)
if [ ${#MICROSERVICES_CHANGED[@]} -gt 0 ]; then
    log "Microservice changes detected: ${MICROSERVICES_CHANGED[*]}"

    # Read device configuration to see which services are installed locally
    CONFIG_DIR="$MAIN_APP_DIR/app/config"
    DEVICE_CONFIG="$CONFIG_DIR/device.yaml"

    if [ -f "$DEVICE_CONFIG" ]; then
        # Get locally configured services
        if command -v yq &> /dev/null; then
            LOCAL_SERVICES=$(yq '.device.roles[]' "$DEVICE_CONFIG" 2>/dev/null || echo "")
        else
            # Fallback: try to parse with grep (less reliable but works without yq)
            LOCAL_SERVICES=$(grep -A 100 "roles:" "$DEVICE_CONFIG" 2>/dev/null | grep "^    - " | sed 's/^    - //' || echo "")
        fi

        # Determine docker-compose file location
        COMPOSE_FILE=""
        if [ -f "$MAIN_APP_DIR/docker-compose.yml" ]; then
            COMPOSE_FILE="$MAIN_APP_DIR/docker-compose.yml"
        elif [ -f "$MAIN_APP_DIR/deploy/docker/docker-compose.dual.yml" ]; then
            COMPOSE_FILE="$MAIN_APP_DIR/deploy/docker/docker-compose.dual.yml"
        elif [ -f "$MAIN_APP_DIR/deploy/docker/docker-compose.lb.yml" ]; then
            COMPOSE_FILE="$MAIN_APP_DIR/deploy/docker/docker-compose.lb.yml"
        fi

        if [ -n "$COMPOSE_FILE" ] && [ -n "$LOCAL_SERVICES" ]; then
            log "Using docker-compose file: $COMPOSE_FILE"

            for service in "${MICROSERVICES_CHANGED[@]}"; do
                # Check if this service is installed on this device
                if echo "$LOCAL_SERVICES" | grep -q "^${service}$"; then
                    log "Restarting Docker container for service: $service"

                    # Restart specific Docker container
                    if docker compose -f "$COMPOSE_FILE" restart "$service" >> "$LOG_FILE" 2>&1; then
                        log "Successfully restarted $service container"
                    else
                        log "WARNING: Failed to restart $service container - trying docker-compose"
                        # Fallback to docker-compose (older systems)
                        if docker-compose -f "$COMPOSE_FILE" restart "$service" >> "$LOG_FILE" 2>&1; then
                            log "Successfully restarted $service container (using docker-compose)"
                        else
                            log "WARNING: Failed to restart $service container"
                        fi
                    fi
                else
                    log "Skipping $service (not installed on this device)"
                fi
            done
        else
            if [ -z "$COMPOSE_FILE" ]; then
                log "WARNING: No docker-compose file found, skipping microservice restarts"
            fi
            if [ -z "$LOCAL_SERVICES" ]; then
                log "WARNING: No local services found in device.yaml, skipping microservice restarts"
            fi
        fi
    else
        log "WARNING: device.yaml not found at $DEVICE_CONFIG, skipping microservice restarts"
    fi
fi

# Handle sketch changes (compile and upload)
if [ "$SKETCH_CHANGED" = true ]; then
    log "Sketch files changed - compiling and uploading..."

    # Compile and upload sketch using native script
    # Note: Docker app is managed by Arduino App Framework, no need to stop/start service
    if [ -f "$MAIN_APP_DIR/scripts/compile_and_upload_sketch.sh" ]; then
        log "Running sketch compilation script..."
        if bash "$MAIN_APP_DIR/scripts/compile_and_upload_sketch.sh" >> "$LOG_FILE" 2>&1; then
            log "Sketch compiled and uploaded successfully"
        else
            log "ERROR: Sketch compilation/upload failed - check logs"
        fi
    else
        log "WARNING: compile_and_upload_sketch.sh not found"
    fi

    # Docker app will automatically reconnect after sketch upload
    log "Sketch upload complete - Docker app will reconnect automatically"
fi

log "Deployment complete"
exit 0
