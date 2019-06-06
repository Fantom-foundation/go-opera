#!/usr/bin/env bash
cd $(dirname $0)

. ./params.sh

CONF=blockade.yaml

cat << HEADER > $CONF
network:                                                                                                                                                                                                    
  name: lachesis
  driver: udn                                                                                                                                                                                               

containers:

HEADER

for i in $(seq $N)
do
    j=$((i % N + 1)) # ring
    cat << NODE >> $CONF
  node_$i:
    image: pos-lachesis-blockade:latest
    container_name: node_$i
    command: start --network=fake:$i/$N --db=/tmp $DSN --peer=node_$j
    expose:
      - "55555"
    deploy:
      resources:
        limits:
          cpus: ${limit_cpu}
          blkio-weight: ${limit_io}

NODE
done

