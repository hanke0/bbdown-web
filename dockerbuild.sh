#!/bin/bash

set -e
set -o pipefail

tag=latest

usage() {
    cat <<EOF
Usage: $0 [OPTION]...

OPTION:
    -h, --help           print this text and exit.
    -t, --tag=NAME       image tag. default to latest.
    -p, --push           push image
EOF
}

get_tag() {
    sed -n -E 's/^LABEL version="(.*)"/\1/p' ./Dockerfile
}

while [ $# -gt 0 ]; do
    case "$1" in
    -h | --help)
        usage
        shift
        ;;
    -t | --tag)
        tag="$2"
        shift 2
        ;;
    -p | --push)
        PUSH=true
        shift
        ;;
    *)
        echo >&2 "bad option: $"
        exit 1
        ;;
    esac
done

cd "$(dirname "$0")"

vtag="$(get_tag)"
image="docker.io/googletranslate/bbdown-web"
docker build -t "${image}:${tag}" .
echo "build ${image}:${tag}"
if [ -n "$vtag" ]; then
    docker build -t "${image}:${vtag}" .
    echo "build ${image}:${vtag}"
fi
if [ "${PUSH}" = true ]; then
    docker push "${image}:${tag}"
    echo "push ${image}:${tag}"
    if [ -n "$tag" ]; then
        docker push "${image}:${vtag}"
        echo "push ${image}:${vtag}"
    fi
fi
