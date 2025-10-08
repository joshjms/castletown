#!/usr/bin/env bash

set -euo pipefail

if [ -d "/tmp/castletown/images/gcc-15-bookworm" ]; then
    exit 0
fi

skopeo copy docker://gcc:15-bookworm oci:/tmp/_tmp_gcc:15-bookworm

umoci raw unpack --rootless \
    --image /tmp/_tmp_gcc:15-bookworm \
    /tmp/_tmp_gcc_15-bookworm

mkdir -p /tmp/castletown/images/gcc-15-bookworm
cp -r /tmp/_tmp_gcc_15-bookworm/* /tmp/castletown/images/gcc-15-bookworm
rm -rf /tmp/_tmp_gcc_15-bookworm

mkdir -p /tmp/castletown/images/gcc-15-bookworm/box
