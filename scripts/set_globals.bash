#!/usr/bin/env bash -euo pipefail

export IFS=$'\n\t'

export PROJECT="${PROJECT:-lachesis}"
export BUILD_DIR="${BUILD_DIR:-$DIR}"
export PEERS_DIR="${PEERS_DIR:-$BUILD_DIR}"
export TARGET_OS="${TARGET_OS:-linux}"
