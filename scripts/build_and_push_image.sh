#!/bin/bash
set -e

# specific registry and image name
REGISTRY="ghcr.io"
USER="gjcourt"
PROJECT="soundbyte"
DATE=$(date +%Y-%m-%d)
TAG="$DATE"

# Function to check if image tag exists
image_exists() {
    docker manifest inspect "$REGISTRY/$USER/$PROJECT:$1" > /dev/null 2>&1
}

# Check if base tag exists
if image_exists "$TAG"; then
    echo "Tag $TAG exists. Finding next available version suffix..."
    SUFFIX=2
    while image_exists "$TAG-v$SUFFIX"; do
        SUFFIX=$((SUFFIX+1))
    done
    TAG="$TAG-v$SUFFIX"
fi

IMAGE="$REGISTRY/$USER/$PROJECT:$TAG"

echo "Building image: $IMAGE"
docker build -t "$IMAGE" .

echo "Pushing image: $IMAGE"
docker push "$IMAGE"

echo "Done! Image pushed to $IMAGE"
