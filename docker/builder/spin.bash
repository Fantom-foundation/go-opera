#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "$DIR/set_globals.bash"

node_num="$1"
ip="$2"
container="$PROJECT$node_num"
docker create -e node_num="$node_num" -e node_addr="$ip" --hostname "$container" --name "$container" --network "$PROJECT-net" --ip "$ip" "$PROJECT"
# docker start "$container"
# docker logs -f "$container"
