#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Cleanup
rm -rf "$DIR/nodes" "$DIR/peers.json"
nodes=$(docker ps -a --no-trunc --filter name='^/lachesis' --format '{{.Names}}')
docker stop -f "$nodes"
docker rm "$nodes"

# Config
n="${n:-1000}"
ip_start="${ip_start:-192.168.0.2}"
subnet="${subnet:-16}"
ip_range="$ip_start/$subnet"

# Run
batch-ethkey -dir nodes -network "$ip_start" -n "$n" > peers.json
docker build --compress --squash --force-rm --tag lachesis "$DIR"
"$DIR/network.bash" "$ip_range"
"$DIR/spin_multi.bash" "$n"

docker start $(docker ps -a --no-trunc --filter name='^/lachesis' --format '{{.Names}}')
