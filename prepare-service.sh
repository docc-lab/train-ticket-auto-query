#!/bin/bash

# Check if all required arguments are provided
if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <service-name> <burst-threshold> <burst-count> <tag-name>"
    echo "Example: $0 ts-cancel-service 10 5 v1.0.0"
    exit 1
fi

# Assign arguments to variables
SERVICE_NAME=$1
BURST_THRESHOLD=$2
BURST_COUNT=$3
TAG_NAME=$4

# Parse the service name to extract the 'xxx' part
SERVICE_PART=$(echo $SERVICE_NAME | sed 's/ts-\(.*\)-service/\1/')

# Capitalize the first letter of SERVICE_PART for the controller name
CONTROLLER_NAME="$(tr '[:lower:]' '[:upper:]' <<< ${SERVICE_PART:0:1})${SERVICE_PART:1}Controller.java"

# Construct the path
FILE_PATH="${SERVICE_NAME}/src/main/java/${SERVICE_PART}/controller/${CONTROLLER_NAME}"

# Navigate to the train-ticket directory
cd /local/train-ticket || exit
sudo chown -R geniuser .

# Switch to the correct branch
sudo git fetch origin '+refs/heads/*:refs/remotes/origin/*'
sudo git switch -c cacti-exp origin/cacti-exp

# Use sed to replace BURST_THRESHOLD and BURST_COUNT in the controller file
sed -i "s/private static final int BURST_THRESHOLD = [0-9]\+;/private static final int BURST_THRESHOLD = ${BURST_THRESHOLD};/" "$FILE_PATH"
sed -i "s/private static final int BURST_COUNT = [0-9]\+;/private static final int BURST_COUNT = ${BURST_COUNT};/" "$FILE_PATH"
echo "Updatec bursty variables"

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

echo "Deployment of ${SERVICE_NAME} with burst threshold ${BURST_THRESHOLD}, burst count ${BURST_COUNT}, and tag ${TAG_NAME} completed."
