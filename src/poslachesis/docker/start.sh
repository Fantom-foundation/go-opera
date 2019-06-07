#!/usr/bin/env bash
cd $(dirname $0)

. ./params.sh

docker network create lachesis

name=pos-lachesis-node

for i in $(seq $N)
do
    j=$((i % N + 1)) # ring
    docker run -d --rm \
		${limits} --net=lachesis --name=$name-$i \
		"pos-lachesis" \
		--network=fake:$i/$N --db=/tmp --peer=$name-$j $DSN \
		start
done

export name

# Prometheus
declare -rl PROMETHEUS="${PROMETHEUS:-no}"
if [ "$PROMETHEUS" == "yes" ]; then
    ./prometheus.sh && make prometheus && make prometheus-on
fi

# Note: We need to know IP of containers before start Prometheus.
