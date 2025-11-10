#!/usr/bin/env bash
set -euo pipefail

log() {
  echo "[castletown-entrypoint] $*"
}

: "${BLOB_ROOT:=/var/castletown/blobs}"
: "${WORK_ROOT:=/tmp/castletown/work}"
: "${IMAGES_DIR:=/var/castletown/images}"
: "${OVERLAYFS_DIR:=/tmp/castletown/overlayfs}"
: "${STORAGE_DIR:=/tmp/castletown/storage}"
: "${LIBCONTAINER_DIR:=/tmp/castletown/libcontainer}"
: "${ROOTFS_DIR:=/tmp/castletown/rootfs}"
: "${PROBLEM_CACHE_DIR:=/var/castletown/problems}"
: "${CASTLETOWN_SKIP_ROOTFS:=1}"

mkdir -p \
  "${BLOB_ROOT}" \
  "${WORK_ROOT}" \
  "${IMAGES_DIR}" \
  "${OVERLAYFS_DIR}" \
  "${STORAGE_DIR}" \
  "${LIBCONTAINER_DIR}" \
  "${ROOTFS_DIR}" \
  "${PROBLEM_CACHE_DIR}"

if [[ "${CASTLETOWN_SKIP_ROOTFS}" != "1" ]]; then
  if ! command -v skopeo >/dev/null 2>&1 || ! command -v umoci >/dev/null 2>&1; then
    log "rootfs bootstrap requested but skopeo/umoci are not installed in this image"
    exit 1
  fi
  if [[ ! -d "${IMAGES_DIR}/gcc-15-bookworm" ]]; then
    log "bootstraping gcc-15-bookworm rootfs into ${IMAGES_DIR}"
    IMAGES_DIR="${IMAGES_DIR}" /opt/castletown/scripts/rootfs.sh
  else
    log "rootfs already present in ${IMAGES_DIR}, skipping bootstrap"
  fi
else
  log "CASTLETOWN_SKIP_ROOTFS=1, skipping rootfs bootstrap"
fi

log "starting castletown: $*"
exec "$@"
