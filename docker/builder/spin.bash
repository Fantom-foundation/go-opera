#!/usr/bin/env bash

node_num="$1"
ip="$2"
container="lachesis$node_num"
docker create -e node_num="$node_num" -e node_addr="$ip" --name "$container" --network lachesis-net --ip "$ip" --rm lachesis
#docker start "$container"
#docker logs -f "$container"
