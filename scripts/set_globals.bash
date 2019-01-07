#!/usr/bin/env bash -euo pipefail

export IFS=$'\n\t'

export PROJECT="${PROJECT:-lachesis}"
export BUILD_DIR="${BUILD_DIR:-$DIR}"
export DATAL_DIR="${DATAL_DIR:-$BUILD_DIR}"
export PEERS_DIR="${PEERS_DIR:-$BUILD_DIR}"
export TARGET_OS="${TARGET_OS:-linux}"
export PATH="$GOPATH/bin:/usr/local/go/bin:$PATH"
