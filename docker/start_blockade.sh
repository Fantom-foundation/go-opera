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
    cat << NODE >> $CONF
  node_$i:
    image: lachesis:latest
    container_name: ${NAME}-$i
    command: --fakenet $i/$N --rpc --rpcaddr 0.0.0.0 --rpcport 18545 --rpccorsdomain "*" --rpcapi "eth,admin,web3" --nousb --metrics
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

