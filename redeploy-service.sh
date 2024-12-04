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

# Initialize array for tracking failed services
failed_services=()

# Function to clean up any local changes and update to latest remote version
cleanup_and_update() {
    log_info "Starting repository cleanup and update"
    log_info "Starting repository cleanup and update"
    echo "Cleaning up local changes..."
#    git restore . 2>/dev/null || true
#    git clean -fd 2>/dev/null || true

    echo "Switching to cacti-exp branch..."
    git switch cacti-exp

    echo "Pulling latest changes..."
    git pull origin cacti-exp --ff-only

    if [ $? -ne 0 ]; then
        log_error "Failed to update to latest version. Exiting."
        log_error "Failed to update to latest version. Exiting."
        exit 1
    fi
    log_success "Repository cleanup and update completed"
}

# Function to build and push Docker image
build_and_push_docker() {
    local service=$1
    local tag=$2

    log_info "Building and pushing Docker image for $service"

    # Navigate to the service directory
    cd "${service}" || { log_error "Failed to navigate to ${service} directory"; return 1; }

    # Build Docker image
    log_info "Building Docker image for $service:$tag"
    if ! sudo docker build -t "docclabgroup/${service}:${tag}" .; then
        log_error "Docker build failed for $service"
        cd ..
        return 1
    fi

    # Push Docker image
    log_info "Pushing Docker image for $service:$tag"
    if ! sudo docker push "docclabgroup/${service}:${tag}"; then
        log_error "Docker push failed for $service"
        cd ..
        return 1
    fi

    cd .. || { log_error "Failed to navigate back from ${service} directory"; return 1; }
    log_success "Successfully built and pushed Docker image for $service"
}

# Function to check rollout status
check_rollout() {
    local service=$1
    local failed_rollouts=()
    local success=true

    log_info "Checking rollout status for updated service"

    log_info "Checking rollout status for $service"
    if ! kubectl rollout status "deployment/${service}" --timeout=300s; then
        log_error "Rollout failed for $service"
        failed_rollouts+=("$service")
        success=false
    else
        log_success "Rollout completed successfully for $service"
    fi

    if [ "$success" = false ]; then
        log_error "The following services failed to roll out: ${failed_rollouts[*]}"
        return 1
    fi

    log_success "Updated service rolled out successfully"
    return 0
}

# Initial cleanup and update
cleanup_and_update

# Main execution starts here
log_info "Starting main execution"
cd /local/train-ticket || { log_error "Failed to navigate to train-ticket directory"; exit 1; }
sudo chown -R $(whoami) .

# First Phase: Maven Build
log_info "Phase 1: Building project with Maven"
if ! mvn clean install -DskipTests; then
    log_error "Maven build failed"
    exit 1
fi
log_success "Maven build completed"

# Second Phase: Build and Push Docker Images
log_info "Phase 2: Building and pushing Docker images"
echo -e "\n----------------------------------------"
if ! build_and_push_docker "$1" "$2"; then
    failed_services+=("$1")
fi
echo "----------------------------------------"

if [ ${#failed_services[@]} -ne 0 ]; then
    log_error "Failed to build/push the following services: ${failed_services[*]}"
    exit 1
fi

# Third Phase: Update Kubernetes Deployments
log_info "Phase 3: Updating Kubernetes deployments"
echo "Updating all service images simultaneously..."

log_info "Setting new image for $1"
kubectl set image "deployment/$1" "$1=docclabgroup/$1:$2" &

# Wait for all kubectl set image commands to complete
log_info "Waiting for all image updates to complete"
wait
log_success "All deployment image updates initiated"

# Fourth Phase: Check Rollout Status
log_info "Phase 4: Checking rollout status for updated service"
if ! check_rollout "$1"; then
    log_error "Deployment failed to roll out properly"
    exit 1
fi


