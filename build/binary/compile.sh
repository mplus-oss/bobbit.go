#!/usr/bin/env bash
set -exu
mkdir -p build/dist

export CGO_ENABLED=1
export CC=gcc

for cmd in ./cmd/*; do
    go get -C "$cmd"

    ldflags="-w -s -linkmode external -extldflags '-static'"
    output="$(basename "$cmd")-$(go env GOOS)-$(go env GOARCH)"

    go build -C "$cmd" -ldflags="$ldflags" -o "../../build/dist/$output"
    echo "$output done"
done
