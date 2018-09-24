#!/usr/bin/env bash

PROJECT="${PROJECT:-lachesis}"
node_num="$1"
printf -v node_num "%03d" "$node_num"
ip="$2"
container="$PROJECT$node_num"
docker create -e node_num="$node_num" -e node_addr="$ip" --hostname "$container" --name "$container" --network "$PROJECT-net" --ip "$ip" --rm "$PROJECT"
#docker start "$container"
#docker logs -f "$container"
