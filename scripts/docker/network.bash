#!/usr/bin/env bash

set -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
. "${DIR%/*}/set_globals.bash"

declare -r name="$PROJECT-net"
docker network rm "$name"
docker network create --driver=bridge --subnet="$1" "$name"
