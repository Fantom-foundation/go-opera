#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "$DIR/set_globals.bash"

rm -rf "$BUILD_DIR/nodes" "$BUILD_DIR/peers.json"
containers="$(docker ps -a --no-trunc --filter name='^/'"$PROJECT" --format '{{.Names}}')"

if [ ! -z "$containers" ]; then
  printf "Stopping & removing $PROJECT containers\n"
  docker kill $containers
  docker rm $containers
fi
