#!/bin/bash

# Function for consistent log formatting
log_info() {
    echo -e "\n[INFO] $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "\n[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - $1" >&2
}

log_success() {
    echo -e "\n[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check if all required arguments are provided
if [ "$#" -ne 5 ]; then
    log_error "Missing required arguments"
    echo "Prerequiste: 1. install maven, 2.login to dockerhub use \"docker login\", 3. make sure train-ticket repo is already in cacti-exp branch"
    echo "Usage: $0 <target-service-name> <bursty-period> <burst-rate> <burst-duration> <tag-name>"
    echo "Example: $0 ts-cancel-service 60 5 10 v1.0.0"
    echo "Parameters:"
    echo "  - bursty-period: Time between bursts in seconds (e.g., 60 for 1 minute)"
    echo "  - burst-rate: Requests per second during burst (e.g., 5)"
    echo "  - burst-duration: Duration of each burst in seconds (e.g., 10)"
    exit 1
fi

# List of all bursty services
BURSTY_SERVICES=(
    "ts-cancel-service"
    "ts-basic-service"
    "ts-travel-service"
    "ts-preserve-service"
)

# Assign arguments to variables
TARGET_SERVICE=$1
BURSTY_PERIOD_SECONDS=$2
BURST_REQUESTS_PER_SEC=$3
BURST_DURATION_SECONDS=$4
TAG_NAME=$5

log_info "Starting service update process with following parameters:"
echo "Target Service: $TARGET_SERVICE"
echo "Bursty Period: $BURSTY_PERIOD_SECONDS seconds"
echo "Burst Rate: $BURST_REQUESTS_PER_SEC requests/second"
echo "Burst Duration: $BURST_DURATION_SECONDS seconds"
echo "Tag Name: $TAG_NAME"

# Function to get controller path based on service name
get_controller_path() {
    local service=$1
    local service_part=$(echo $service | sed 's/ts-\(.*\)-service/\1/')
    local controller_name="$(tr '[:lower:]' '[:upper:]' <<< ${service_part:0:1})${service_part:1}Controller.java"
    
    # Check if it's basic service which has different path
    if [ "$service" = "ts-basic-service" ]; then
        echo "${service}/src/main/java/fdse/microservice/controller/${controller_name}"
    else
        echo "${service}/src/main/java/${service_part}/controller/${controller_name}"
    fi
}

# Function to clean up any local changes and update to latest remote version
cleanup_and_update() {
    log_info "Starting repository cleanup and update"
    echo "Cleaning up local changes..."
    git restore . 2>/dev/null || true
    git clean -fd 2>/dev/null || true
    
    echo "Switching to exp-dev branch..."
    git switch exp-dev
    
    echo "Pulling latest changes..."
    git pull origin exp-dev --ff-only
    
    if [ $? -ne 0 ]; then
        log_error "Failed to update to latest version. Exiting."
        exit 1
    fi
    log_success "Repository cleanup and update completed"
}

# Function to update burst parameters in a service
update_service_params() {
    local service=$1
    local period=$2
    local rate=$3
    local duration=$4
    
    log_info "Updating parameters for service: $service"
    local file_path=$(get_controller_path "$service")
    
    if [ ! -f "$file_path" ]; then
        log_error "Controller file not found at $file_path"
        return 1
    fi
    
    echo "Controller file path: $file_path"
    echo "Applying changes:"
    echo "  - BURSTY_PERIOD_SECONDS: $period"
    echo "  - BURST_REQUESTS_PER_SEC: $rate"
    echo "  - BURST_DURATION_SECONDS: $duration"
    
    # Create backup of original file
    cp "$file_path" "${file_path}.bak"
    
    # Apply changes and verify
    sed -i "s/private static final int BURSTY_PERIOD_SECONDS = [0-9]\+;/private static final int BURSTY_PERIOD_SECONDS = ${period};/" "$file_path"
    sed -i "s/private static final int BURST_REQUESTS_PER_SEC = [0-9]\+;/private static final int BURST_REQUESTS_PER_SEC = ${rate};/" "$file_path"
    sed -i "s/private static final int BURST_DURATION_SECONDS = [0-9]\+;/private static final int BURST_DURATION_SECONDS = ${duration};/" "$file_path"
    
    # Verify changes were applied
    if grep -q "BURSTY_PERIOD_SECONDS = ${period}" "$file_path" && \
       grep -q "BURST_REQUESTS_PER_SEC = ${rate}" "$file_path" && \
       grep -q "BURST_DURATION_SECONDS = ${duration}" "$file_path"; then
        log_success "Successfully updated parameters for $service"
        rm "${file_path}.bak"
    else
        log_error "Failed to update all parameters for $service"
        log_info "Restoring backup file"
        mv "${file_path}.bak" "$file_path"
        return 1
    fi
}

# Function to build and deploy a service
build_and_deploy_service() {
    local service=$1
    local tag=$2

    log_info "Starting deployment process for $service"

    # Navigate to the service directory
    cd "${service}" || { log_error "Failed to navigate to ${service} directory"; return 1; }

    # Build Docker image
    log_info "Building Docker image for $service:$tag"
    if ! sudo docker build -t "docclabgroup/${service}:${tag}" .; then
        log_error "Docker build failed for $service"
        return 1
    fi

    # Push Docker image
    log_info "Pushing Docker image for $service:$tag"
    if ! sudo docker push "docclabgroup/${service}:${tag}"; then
        log_error "Docker push failed for $service"
        return 1
    fi

    # Update Kubernetes deployment
    log_info "Updating Kubernetes deployment for $service"
    if ! kubectl set image "deployment/${service}" "${service}=docclabgroup/${service}:${tag}"; then
        log_error "Kubernetes deployment update failed for $service"
        return 1
    fi

    # Wait for rollout
    log_info "Waiting for rollout completion of $service"
    if ! kubectl rollout status "deployment/${service}"; then
        log_error "Kubernetes rollout failed for $service"
        return 1
    fi

    cd .. || { log_error "Failed to navigate back from ${service} directory"; return 1; }
    log_success "Successfully deployed $service"
}

# Main execution starts here
log_info "Starting main execution"
cd /local/train-ticket || { log_error "Failed to navigate to train-ticket directory"; exit 1; }
sudo chown -R $(whoami) .

# Initial cleanup and update
cleanup_and_update

# First Phase: Update Parameters
log_info "Phase 1: Updating service parameters"
echo "Total services to process: ${#BURSTY_SERVICES[@]}"

for service in "${BURSTY_SERVICES[@]}"; do
    echo -e "\n----------------------------------------"
    log_info "Updating parameters for service: $service"

    if [ "$service" = "$TARGET_SERVICE" ]; then
        log_info "$service is the target service - applying specified burst parameters"
        update_service_params "$service" "$BURSTY_PERIOD_SECONDS" "$BURST_REQUESTS_PER_SEC" "$BURST_DURATION_SECONDS"
    else
        log_info "$service is not the target service - setting zero burst parameters"
        update_service_params "$service" "0" "0" "0"
    fi
    echo "----------------------------------------"
done

# Second Phase: Maven Build
log_info "Phase 2: Building project with Maven"
if ! mvn clean install -DskipTests; then
    log_error "Maven build failed"
    exit 1
fi
log_success "Maven build completed"

# Third Phase: Build and Deploy Services
log_info "Phase 3: Building and deploying services"
for service in "${BURSTY_SERVICES[@]}"; do
    echo -e "\n----------------------------------------"
    log_info "Building and deploying service: $service"
    build_and_deploy_service "$service" "$TAG_NAME"
    echo "----------------------------------------"
done

log_success "All deployments completed successfully!"
echo -e "\nFinal Configuration Summary:"
echo "Target Service: ${TARGET_SERVICE}"
echo "Target Configuration:"
echo "- Bursty Period: ${BURSTY_PERIOD_SECONDS} seconds"
echo "- Burst Rate: ${BURST_REQUESTS_PER_SEC} requests per second"
echo "- Burst Duration: ${BURST_DURATION_SECONDS} seconds"
echo "- Image Tag: ${TAG_NAME}"
echo "All other services have been set to zero burst parameters"