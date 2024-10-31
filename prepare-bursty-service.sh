#!/bin/bash

# Check if all required arguments are provided
if [ "$#" -ne 5 ]; then
    echo "Prerequiste: 1. install maven, 2.login to dockerhub use "docker login", 3. make sure train-ticket repo is already in cacti-exp branch"
    echo "Usage: $0 <service-name> <bursty-period> <burst-rate> <burst-duration> <tag-name>"
    echo "Example: $0 ts-cancel-service 60 5 10 v1.0.0"
    echo "Parameters:"
    echo "  - bursty-period: Time between bursts in seconds (e.g., 60 for 1 minute)"
    echo "  - burst-rate: Requests per second during burst (e.g., 5)"
    echo "  - burst-duration: Duration of each burst in seconds (e.g., 10)"
    exit 1
fi

# Assign arguments to variables
SERVICE_NAME=$1
BURSTY_PERIOD_SECONDS=$2
BURST_REQUESTS_PER_SEC=$3
BURST_DURATION_SECONDS=$4
TAG_NAME=$5

# Parse the service name to extract the 'xxx' part
SERVICE_PART=$(echo $SERVICE_NAME | sed 's/ts-\(.*\)-service/\1/')

# Capitalize the first letter of SERVICE_PART for the controller name
CONTROLLER_NAME="$(tr '[:lower:]' '[:upper:]' <<< ${SERVICE_PART:0:1})${SERVICE_PART:1}Controller.java"

# Construct the path
# FILE_PATH="${SERVICE_NAME}/src/main/java/${SERVICE_PART}/controller/${CONTROLLER_NAME}"
# handle unconsistent dir structure in some service
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

FILE_PATH=$(get_controller_path "$SERVICE_NAME")

# Navigate to the train-ticket directory
cd /local/train-ticket || exit
sudo chown -R $(whoami) .

# Switch to the correct branch
git switch exp-dev
git pull origin exp-dev

# Use sed to replace burst parameters in the controller file
sed -i "s/private static final int BURSTY_PERIOD_SECONDS = [0-9]\+;/private static final int BURSTY_PERIOD_SECONDS = ${BURSTY_PERIOD_SECONDS};/" "$FILE_PATH"
sed -i "s/private static final int BURST_REQUESTS_PER_SEC = [0-9]\+;/private static final int BURST_REQUESTS_PER_SEC = ${BURST_REQUESTS_PER_SEC};/" "$FILE_PATH"
sed -i "s/private static final int BURST_DURATION_SECONDS = [0-9]\+;/private static final int BURST_DURATION_SECONDS = ${BURST_DURATION_SECONDS};/" "$FILE_PATH"
echo "Updated bursty load variables:"
echo "- Bursty Period: ${BURSTY_PERIOD_SECONDS}s"
echo "- Burst Rate: ${BURST_REQUESTS_PER_SEC} req/s" 
echo "- Burst Duration: ${BURST_DURATION_SECONDS}s"

# Build the project
mvn clean install -DskipTests

# Navigate to the service directory
cd "${SERVICE_NAME}" || exit

# Build and push the Docker image
docker build -t "docclabgroup/${SERVICE_NAME}:${TAG_NAME}" .
docker push "docclabgroup/${SERVICE_NAME}:${TAG_NAME}"

# Update the Kubernetes deployment
kubectl set image "deployment/${SERVICE_NAME}" "${SERVICE_NAME}=docclabgroup/${SERVICE_NAME}:${TAG_NAME}"

# Wait for the rollout to complete
kubectl rollout status "deployment/${SERVICE_NAME}"

echo "Deployment completed successfully!"
echo "Service: ${SERVICE_NAME}"
echo "Configuration:"
echo "- Bursty Period: Every ${BURSTY_PERIOD_SECONDS} seconds"
echo "- Burst Rate: ${BURST_REQUESTS_PER_SEC} requests per second"
echo "- Burst Duration: ${BURST_DURATION_SECONDS} seconds"
echo "- Image Tag: ${TAG_NAME}"