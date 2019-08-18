#!/usr/bin/env bash
cd $(dirname $0)

. ./_params.sh

CONF=blockade.yaml

cat << HEADER > $CONF
network:                                                                                                                                                                                                    
  driver: udn                                                                                                                                                                                               

containers:
HEADER

for i in $(seq $N)
do
    j=$((i % N + 1)) # ring
    cat << NODE >> $CONF
  node_$i:
    image: pos-lachesis-blockade:latest
    container_name: ${NAME}-$i
    command: start --network=fake:$i/$N --db=/tmp --peer=${NAME}-$j --metrics
    expose:
      - "55555"
    deploy:
      resources:
        limits:
          cpus: ${LIMIT_CPU}
          blkio-weight: ${LIMIT_IO}
NODE
done

blockade up

NETWORK=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}' ${NAME}-1)

. ./_prometheus.sh

