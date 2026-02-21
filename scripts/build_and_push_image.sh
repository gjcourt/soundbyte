#!/bin/bash
set -e

# Configurable via environment variables (set by Makefile or manually)
REGISTRY="${REGISTRY:-ghcr.io}"
USER="${REGISTRY_USER:-gjcourt}"
PROJECT="soundbyte"
TAG="${IMAGE_TAG:-$(date +%Y-%m-%d)}"
PLATFORM="${PLATFORM:-linux/amd64}"

# Function to check if image tag exists
image_exists() {
    docker manifest inspect "$REGISTRY/$USER/$PROJECT:$1" > /dev/null 2>&1
}

# Check if base tag exists and find next available suffix
if image_exists "$TAG"; then
    echo "Tag $TAG exists. Finding next available version suffix..."
    SUFFIX=2
    while image_exists "$TAG-v$SUFFIX"; do
        SUFFIX=$((SUFFIX+1))
    done
    TAG="$TAG-v$SUFFIX"
fi

IMAGE="$REGISTRY/$USER/$PROJECT:$TAG"
PLATFORM=${PLATFORM:-"linux/amd64,linux/arm64"}

echo "Building and pushing image: $IMAGE for platforms: $PLATFORM"
docker buildx build --platform "$PLATFORM" -t "$IMAGE" --push .

echo "Done! Image pushed to $IMAGE"
