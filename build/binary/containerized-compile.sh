#!/usr/bin/env bash
set -exu

platforms=("linux/amd64" "linux/arm64")

if command -v docker &> /dev/null
then
    CONTAINER_RUNTIME="docker"
elif command -v podman &> /dev/null
then
    CONTAINER_RUNTIME="podman"
else
    echo "CONTAINER_RUNTIME not found"
    exit 1
fi

for platform in "${platforms[@]}"; do
    "$CONTAINER_RUNTIME" build \
        --platform "$platform" \
        --output type=local,dest=./build/dist \
        --file ./build/binary/build.Dockerfile \
        .
done
