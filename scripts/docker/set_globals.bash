#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export PROJECT="${PROJECT:-lachesis}"
export BUILD_DIR="${BUILD_DIR:-$DIR}"
export PEERS_DIR="${PEERS_DIR:-$BUILD_DIR}"
