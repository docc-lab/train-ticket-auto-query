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

cd /local/train-ticket-2 || { log_error "Failed to navigate to train-ticket directory"; exit 1; }
sudo chown -R $(whoami) .

# First Phase: Maven Build
log_info "Phase 2: Building project with Maven"
if ! mvn clean install -DskipTests; then
    log_error "Maven build failed"
    exit 1
fi
log_success "Maven build completed"

# Second Phase: Build and Deploy Services
echo -e "\n----------------------------------------"
log_info "Building and deploying service: $1"
build_and_deploy_service "$1" "$2"
echo "----------------------------------------"


