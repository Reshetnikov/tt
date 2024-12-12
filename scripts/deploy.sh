#!/bin/bash

#The script is run from the development host.
#The script creates an image in ECR, a task definition and a service in ECS. With tags and a port corresponding to the GIT tag.

AWS_REGION="us-east-1"
ECR_REPO="010928182544.dkr.ecr.us-east-1.amazonaws.com/tt/server"
CLUSTER_NAME="tt-cluster"
TASK_DEFINITION_NAME="definition"
SERVICE_NAME="service"



GIT_TAG=$(git describe --tags --abbrev=0 2>/dev/null)
if [ -z "$GIT_TAG" ]; then
  echo "Error: No Git tag found. Please create a tag and try again."
  exit 1
fi
VERSION="${VERSION:-$GIT_TAG}"
VERSION_NUMBER=$(echo "$VERSION" | grep -oE '[0-9]+$')
if [ -z "$VERSION_NUMBER" ]; then
  echo "Error: No numeric version found in tag '$VERSION'."
  exit 1
fi
echo "Using version: $VERSION"
echo "Extracted version number: $VERSION_NUMBER"