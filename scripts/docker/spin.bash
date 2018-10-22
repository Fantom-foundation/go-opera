#!/usr/bin/env bash

set -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "${DIR%/*}/set_globals.bash"

"$DIR/install_deps.bash"

declare -ri node_num="$1"
declare -r ip="$2"
declare -r container="$PROJECT$node_num"
#env_vars="$ENV_VARS"
#env_vars="${env_vars//$'\n'/}"
#docker create "$env_vars" --hostname "$container" --name "$container" --network "$PROJECT-net" --ip "$ip" "$PROJECT"
docker create -e node_num="$node_num" -e node_addr="$ip" --hostname "$container" --name "$container" --network "$PROJECT-net" --ip "$ip" "$PROJECT"
# docker start "$container"
# docker logs -f "$container"
