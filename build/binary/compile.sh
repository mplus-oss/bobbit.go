#!/usr/bin/env bash
set -exu
mkdir -p build/dist

arch_build=$(cat build/arch-list)

for cmd in ./cmd/*; do
    go get -C "$cmd"

    for arch in ${arch_build[@]}; do
        export GOOS="${arch%/*}"
        export GOARCH="${arch#*/}"

        ldflags="-w -s"
        output="$(basename "$cmd")-$GOOS-$GOARCH"

        go build -C "$cmd" -ldflags="$ldflags" -o "../../build/dist/$output"
        echo "$output done"
    done
done
