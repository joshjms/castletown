#!/usr/bin/env bash

set -euo pipefail

if [ -d "/tmp/_tmp_gcc_15-bookworm" ]; then
    exit 0
fi

skopeo copy docker://gcc:15-bookworm oci:/tmp/_tmp_gcc:15-bookworm

umoci raw unpack --rootless \
    --image /tmp/_tmp_gcc:15-bookworm \
    /tmp/_tmp_gcc_15-bookworm
