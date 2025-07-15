#!/usr/bin/env bash
set -exu
container_name=${DOCKER_CONTAINER_NAME:-bobbit-docker}

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

"$CONTAINER_RUNTIME" build \
    -t "$container_name:latest" \
    --file ./build/binary/build.Dockerfile \
    .

"$CONTAINER_RUNTIME" create \
    --name "$container_name" \
    "$container_name:latest" \
    $([ "$CONTAINER_RUNTIME" = 'docker' ] && echo "/bin/true")

"$CONTAINER_RUNTIME" cp \
    $([ "$CONTAINER_RUNTIME" = 'podman' ] && echo "--overwrite") \
    "$container_name:/dist" \
    build/

"$CONTAINER_RUNTIME" rm "$container_name"
"$CONTAINER_RUNTIME" image rm "$container_name:latest"

