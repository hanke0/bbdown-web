#!/bin/bash

set -e

release_platform() {
    while [ $# -gt 0 ]; do
        name="./dist/bbdown-web-$1-$2"
        if [ "$1" = "windows" ]; then
            name="$name.exe"
        fi
        GOOS=$1 GOARCH=$2 go build -o "$name" ./cmd
        shift 2
    done
}

rm -rf ./dist
mkdir -p ./dist

release_platform \
    linux amd64 \
    linux 386 \
    windows amd64 \
    windows 386 \
    darwin arm64 \
    darwin amd64

cd ./dist || exit 1
md5sum >md5.sum ./*
