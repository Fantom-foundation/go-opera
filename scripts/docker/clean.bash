#!/usr/bin/env bash

set -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "${DIR%/*}/set_globals.bash"

rm -rf "$BUILD_DIR/nodes" "$BUILD_DIR/peers.json"

if [ $(pgrep -f docker) ]; then
  exit 0
fi

containers="$(docker ps -a --no-trunc --filter name='^/'"$PROJECT" --format '{{.Names}}')"

if [ ! -z "$containers" ]; then
  printf "Stopping & removing $PROJECT containers\n"
  docker kill -s9 $containers
  docker rm $containers
  docker rmi lachesis
fi
