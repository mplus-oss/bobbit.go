#!/usr/bin/env bash
set -exu
mkdir -p build/dist

GOOS=${GOOS:-linux}
GOARCH=${GOARCH:-amd64}

for cmd in ./cmd/*; do
    go get -C "$cmd"

    ldflags="-w -s"
    output="$(basename "$cmd")-$GOOS-$GOARCH"

    go build -C "$cmd" -ldflags="$ldflags" -o "../../build/dist/$output"
done

