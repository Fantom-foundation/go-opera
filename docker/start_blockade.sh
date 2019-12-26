#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


CONF=blockade.yaml

cat << HEADER > $CONF
network:                                                                                                                                                                                                    
  driver: udn                                                                                                                                                                                               

containers:
HEADER

for ((i=$N-1;i>=0;i-=1))
do
    ACC=$(($i+1)) 

    cat << NODE >> $CONF
  node$i:
    image: lachesis:latest
    container_name: node$i
    command: --fakenet=${ACC}/$N --rpc --rpcaddr="0.0.0.0" --rpcport=18545 --rpccorsdomain="*" --rpcapi="eth,admin,web3,txpool,ftm,sfc" --nousb --metrics
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

. ./_prometheus.sh
