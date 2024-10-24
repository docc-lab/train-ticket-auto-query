# Concurrent Train-Ticket-Auto-Query Fork 

This repo is a fork from [train-ticket's auto-query load generater](https://github.com/FudanSELab/train-ticket-auto-query). We implemented a concurrent version of it in golang under tt-coucurrent-load-generator dir (mainly because of horriable multi-threading in python). We also added a warmup mode for easily populating DB with orders in different states.

## Usage

 1. Find IP address of trainticket ui services

     ```bash
     kubectl get services | grep ts-ui-dashboard
     ```

 2. Populate the DB with warmup mode of concurrent load generator. Call it with `-warmup` flag to find the usage:
     ```bash
    cd /local/train-ticket-auto-query/tt-concurrent-load-generator
    ./tt-concurrent-load-generator -warmup
     ```

 3. Use concurrent load generator. 
    We have been updating the API and functionality of it, so directly call it for usage info.

     ```bash
     cd /local/train-ticket-auto-query/tt-concurrent-load-generator
     ./tt-concurrent-load-generator
     ```


## Script for building and replacing pod image

We also provide script integrating the following steps of building & replacing pod image in trainticket k8s deployment. You may call the `prepare-victime-service.sh` to find usage. It contains several steps for our exp, you might need to remove unnecessary part or modify for your need. 

Forget about the other script, though.

If you don't want to use above script, build new jars with `mvn clean install` under trainticket dir, and then build docker image and replace desired pod with new image:

```bash

# Build and push the Docker image
docker build -t "<your-image-registry>/<your-updated-service>:<your-image-tag>" .
docker push "<your-image-registry>/<your-updated-service>:<your-image-tag>"

# Update the Kubernetes deployment
kubectl set image "deployment/<your-updated-service>" "<your-updated-service>=<your-image-registry>/<your-updated-service>:<your-image-tag>"

# Wait for the rollout to complete
kubectl rollout status "deployment/<your-updated-service>"
```