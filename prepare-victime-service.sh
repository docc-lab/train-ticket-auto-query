if [ "$#" -ne 2 ]; then
    echo "Prerequiste: 1. install maven, 2.login to dockerhub use "docker login", 3. make sure train-ticket repo is already in cacti-exp branch"
    echo "For non-parameter tuning service--just build, push, and replace image."
    echo "Usage: $0 <service-name> <tag-name>"
    echo "Example: $0 ts-order-service v1.0.0"
    exit 1
fi

SERVICE_NAME=$1
TAG_NAME=$2

# Navigate to the train-ticket directory
cd /local/train-ticket || exit
sudo chown -R $(whoami) .

# Switch to the correct branch
git switch cacti-exp
git pull origin cacti-exp

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

echo "Deployment of ${SERVICE_NAME} and tag ${TAG_NAME} completed."