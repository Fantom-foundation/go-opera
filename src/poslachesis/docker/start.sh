#!/usr/bin/env bash

declare -ri n="${n:-3}"

docker network create lachesis

limit_cpu=$(echo "scale=2; 1/$n" | bc)
limit_io=$(echo "500/$n" | bc)
limits="--cpus=${limit_cpu} --blkio-weight=${limit_io}"

for i in $(seq $n)
do
    j=$(($i % $n + 1)) # ring
    docker run -d --rm --name=pos-lachesis-node-$i ${limits} --net=lachesis "pos-lachesis" --fakegen=$i/$n --db=/tmp --peer=pos-lachesis-node-$j
done