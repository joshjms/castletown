#!/usr/bin/env bash

set -euo pipefail

PLATFORMS=("linux/amd64")
VERSION=0.2.0

rm -rf dist
mkdir -p dist

for PLATFORM in "${PLATFORMS[@]}"
do
    IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT="castletown-${GOOS}-${GOARCH}"

    echo "building castletown for ${GOOS}/${GOARCH}"

    GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-X github.com/joshjms/castletown/cmd.Version=${VERSION}" -o "dist/${OUTPUT}" main.go
done

echo "build complete"
