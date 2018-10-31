#!/usr/bin/env bash

set -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
declare -r DOCKER_PID=/var/run/docker.pid

. "${DIR%/*}/set_globals.bash"

rm -rf "$BUILD_DIR/nodes" "$BUILD_DIR/peers.json"

if [ ! -f $DOCKER_PID ] || [ ! $(pgrep --pidfile $DOCKER_PID) ]; then
  exit 0
fi

containers="$(docker ps -a --no-trunc --filter name='^/'"$PROJECT" --format '{{.Names}}')"

if [ ! -z "$containers" ]; then
    printf "Stopping & removing $PROJECT containers\n"
    for container in $containers; do
	if [ $(docker inspect -f '{{.State.Running}}' $container) != "false" ]; then
	    docker kill -s9 $container
	fi
    done
    docker rm $containers
    docker rmi lachesis
fi
