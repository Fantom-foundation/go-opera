#!/usr/bin/env bash

declare -ri n="${n:-3}"
CONF=blockade.yaml

limit_cpu=$(echo "scale=2; 1/$n" | bc)                                                                                                                                                                      
limit_io=$(echo "500/$n" | bc)                                                                                                                                                                              
limits="--cpus=${limit_cpu} --blkio-weight=${limit_io}"                                                                                                                                                     


cat << HEADER > $CONF
network:                                                                                                                                                                                                    
  driver: udn                                                                                                                                                                                               

containers:

HEADER

for i in $(seq $n)
do
    j=$((i % n + 1)) # ring
    cat << NODE >> $CONF
  node_$i:
    image: pos-lachesis-blockade:latest
    container_name: node_$i
    command: start --fakegen=$i/$n --db=/tmp --peer=node_$j
    expose:
      - "55555"
    deploy:
      resources:
        limits:
          cpus: ${limit_cpu}
          blkio-weight: ${limit_io}

NODE
done