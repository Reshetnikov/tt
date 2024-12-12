#!/bin/bash

# The script is run from the development host.
# The script creates an image in ECR, a task definition and a service in ECS. With tags and a port corresponding to the GIT tag.

AWS_REGION="us-east-1"
ECR_REPO="010928182544.dkr.ecr.us-east-1.amazonaws.com/tt/server"
CLUSTER_NAME="tt-cluster"
TASK_DEFINITION_NAME="def" # + "-" + TAG
TASK_DEFINITION_TPL_NAME="def-tpl"
SERVICE_NAME="service"


# Get version tag from Git
TAG=$(git describe --tags --abbrev=0 2>/dev/null)
if [ -z "$TAG" ]; then
  echo "Error: No Git tag found. Please create a tag and try again."
  exit 1
fi
VERSION_NUMBER=$(echo "$TAG" | grep -oE '[0-9]+$')
if [ -z "$VERSION_NUMBER" ]; then
  echo "Error: No numeric version found in tag '$TAG'."
  exit 1
fi
echo "Using version: $TAG"
echo "Extracted version number: $VERSION_NUMBER"

echo -n "$TAG" > web/templates/appVer.txt
echo "Tag saved to web/templates/appVer.txt"

# Build docker image
docker build -t $ECR_REPO:$TAG -f Dockerfile.prod .
# Send Docker image to ECR
echo "Loging to ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_REPO
echo "Pushing to ECR... $ECR_REPO:$TAG" 
docker push $ECR_REPO:$TAG


# Create a new Task Definition
echo "Get the current tpl Task Definition..."
TASK_DEFINITION_TPL=$(aws ecs describe-task-definition \
  --region $AWS_REGION \
  --task-definition $TASK_DEFINITION_TPL_NAME \
  --query 'taskDefinition' \
  --output json)
if [ $? -ne 0 ]; then
  echo "Error: Failed to describe task definition $TASK_DEFINITION_TPL_NAME."
  exit 1
fi

echo "Update the parameters"
NEW_TASK_DEFINITION_NAME="$TASK_DEFINITION_NAME-$TAG"
UPDATED_TASK_DEFINITION=$(echo "$TASK_DEFINITION_TPL" | jq \
  --arg family "$NEW_TASK_DEFINITION_NAME" \
  --arg image "$ECR_REPO:$TAG" \
  --argjson versionNumber $VERSION_NUMBER \
  'del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .compatibilities, .registeredAt, .registeredBy) |
   .family = $family |
   .containerDefinitions[0].image = $image |
   .containerDefinitions[0].portMappings[0].hostPort += $versionNumber')

echo "Registering the updated Task Definition"
echo "$UPDATED_TASK_DEFINITION" > /tmp/updated_task_definition.json
TASK_DEFINITION_ARN=$(aws ecs register-task-definition \
  --region $AWS_REGION \
  --cli-input-json file:///tmp/updated_task_definition.json \
  --query 'taskDefinition.taskDefinitionArn' \
  --output text)
rm /tmp/updated_task_definition.json
if [ $? -ne 0 ]; then
  echo "Error: Failed to updated task definition $NEW_TASK_DEFINITION_NAME."
  exit 1
fi
echo "Task Definition $TASK_DEFINITION_ARN created."

# Create a new Service
SERVICE_NAME="$SERVICE_NAME-$TAG"
aws ecs create-service \
  --region $AWS_REGION \
  --cluster $CLUSTER_NAME \
  --service-name $SERVICE_NAME \
  --task-definition $TASK_DEFINITION_ARN \
  --desired-count 1 \
  --launch-type EC2 \
  --output text
if [ $? -ne 0 ]; then
  echo "Error: Failed to create service $NEW_TASK_DEFINITION_NAME."
  exit 1
fi
echo "Task Definition $TASK_DEFINITION_ARN created."
echo "ECS Service $SERVICE_NAME created whit Task Definition $TASK_DEFINITION_ARN."