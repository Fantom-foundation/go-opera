#!/usr/bin/env bash

set -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
declare -r parent_dir="${DIR%/*}"
declare -r gparent_dir="${parent_dir%/*}"
declare -r DOCKER_PID=/var/run/docker.pid

. "${DIR%/*}/set_globals.bash"


for dir in "$BUILD_DIR" "$DIR" "$parent_dir" "$gparent_dir"; do
  rm -rf "$dir"/{nodes,peers.json,lachesis_d*}
done

if [ ! -f "$DOCKER_PID" ] || [ ! $(pgrep --pidfile "$DOCKER_PID") ]; then
  exit 0
fi

containers="$(docker ps -a --no-trunc --filter name='^/'"$PROJECT" --format '{{.Names}}')"

if [ ! -z "$containers" ]; then
    printf "Stopping & removing $PROJECT containers\n"
    for container in $containers; do
	if [ $(docker inspect -f '{{.State.Running}}' "$container") != "false" ]; then
	    docker kill -s9 "$container"
	fi
    done
    docker rm $containers
    docker rmi "$PROJECT"
fi
