#!/usr/bin/env bash
cd $(dirname $0)

. ./params.sh

docker network create lachesis

for i in $(seq $N)
do
    j=$((i % N + 1)) # ring
    docker run -d --rm \
		${limits} --net=lachesis --name=pos-lachesis-node-$i \
		"pos-lachesis" \
		--network=fake:$i/$N --db=/tmp --peer=pos-lachesis-node-$j $DSN \
		start
done
