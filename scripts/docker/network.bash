#!/usr/bin/env bash

set -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
. "${DIR%/*}/set_globals.bash"

declare -r name="$PROJECT-net"
if [ $(docker network ls -qf name=$name) != "" ]; then
    docker network rm "$name"
fi
docker network create --driver=bridge --subnet="$1" "$name"
