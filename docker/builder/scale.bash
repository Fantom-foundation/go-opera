#!/usr/bin/env bash
rm -rf nodes peers.json

n="${n:-1000}"
ip_start="${ip_start:-192.168.0.5}"
subnet="${subnet:-16}"
ip_range="$ip_start/$subnet"

batch-ethkey -dir nodes -network "$ip_start" -n "$n" > peers.json
./network.bash "$ip_range"
./spin_multi.bash "$n"
docker start $(docker ps -a --no-trunc --filter name='^/lachesis' --format '{{.Names}}')
