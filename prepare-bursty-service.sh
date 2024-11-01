#!/bin/bash

# Check if all required arguments are provided
if [ "$#" -ne 5 ]; then
    echo "Prerequiste: 1. install maven, 2.login to dockerhub use \"docker login\", 3. make sure train-ticket repo is already in cacti-exp branch"
    echo "Functionality: 1. update the inputted service's burstness parameter to the desired, 2. update the remaining bursty services' parameter to all 0"
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
    echo "Cleaning up local changes..."
    git restore . 2>/dev/null || true
    git clean -fd 2>/dev/null || true
    
    echo "Switching to exp-dev branch..."
    git switch exp-dev
    
    echo "Pulling latest changes..."
    git pull origin exp-dev --ff-only
    
    if [ $? -ne 0 ]; then
        echo "Error: Failed to update to latest version. Exiting."
        exit 1
    fi
}

# Function to update burst parameters in a service
update_service_params() {
    local service=$1
    local period=$2
    local rate=$3
    local duration=$4
    
    local file_path=$(get_controller_path "$service")
    
    if [ ! -f "$file_path" ]; then
        echo "Error: Controller file not found at $file_path"
        return 1
    fi
    
    sed -i "s/private static final int BURSTY_PERIOD_SECONDS = [0-9]\+;/private static final int BURSTY_PERIOD_SECONDS = ${period};/" "$file_path"
    sed -i "s/private static final int BURST_REQUESTS_PER_SEC = [0-9]\+;/private static final int BURST_REQUESTS_PER_SEC = ${rate};/" "$file_path"
    sed -i "s/private static final int BURST_DURATION_SECONDS = [0-9]\+;/private static final int BURST_DURATION_SECONDS = ${duration};/" "$file_path"
    
    echo "Updated $service:"
    echo "- Bursty Period: ${period}s"
    echo "- Burst Rate: ${rate} req/s"
    echo "- Burst Duration: ${duration}s"
}

# Function to build and deploy a service
build_and_deploy_service() {
    local service=$1
    local tag=$2
    
    echo "Building and deploying $service..."
    
    # Navigate to the service directory
    cd "${service}" || return 1
    
    # Build and push the Docker image
    docker build -t "docclabgroup/${service}:${tag}" .
    docker push "docclabgroup/${service}:${tag}"
    
    # Update the Kubernetes deployment
    kubectl set image "deployment/${service}" "${service}=docclabgroup/${service}:${tag}"
    
    # Wait for the rollout to complete
    kubectl rollout status "deployment/${service}"
    
    cd .. || return 1
}

# Main execution starts here
cd /local/train-ticket || exit
sudo chown -R $(whoami) .

# Initial cleanup and update
cleanup_and_update

# Build the entire project first
mvn clean install -DskipTests

# Process all services
for service in "${BURSTY_SERVICES[@]}"; do
    if [ "$service" = "$TARGET_SERVICE" ]; then
        # Update target service with specified burst parameters
        update_service_params "$service" "$BURSTY_PERIOD_SECONDS" "$BURST_REQUESTS_PER_SEC" "$BURST_DURATION_SECONDS"
    else
        # Set all other services to zero burst parameters
        update_service_params "$service" "0" "0" "0"
    fi
    
    # Build and deploy each service with the same tag
    build_and_deploy_service "$service" "$TAG_NAME"
done

echo "All deployments completed successfully!"
echo "Target Service: ${TARGET_SERVICE}"
echo "Target Configuration:"
echo "- Bursty Period: ${BURSTY_PERIOD_SECONDS} seconds"
echo "- Burst Rate: ${BURST_REQUESTS_PER_SEC} requests per second"
echo "- Burst Duration: ${BURST_DURATION_SECONDS} seconds"
echo "- Image Tag: ${TAG_NAME}"
echo "All other services set to zero burst parameters"