#!/usr/bin/env bash

node_num="$1"
ip=192.168.0.$(( 2+"$node_num" ))
container="lachesis$node_num"
docker create -e node_num="$node_num" -e node_addr="$ip" --name "$container" --network lachesis-net --ip "$ip" --rm 'lachesis'
docker start "$container"
docker logs -f "$container"
