#!/bin/bash

source $(dirname $0)/set_env.sh


for i in `seq 1 $N`
do

  NAME=node$i
  PORT=$(($PORT_BASE+$i))
  RPCP=$(($RPCP_BASE+$i))

  docker $SWARM service create \
    --name ${NAME} \
    --publish published=${PORT},target=${PORT},mode=ingress,protocol=tcp \
    --publish published=${PORT},target=${PORT},mode=ingress,protocol=udp \
    --publish published=${RPCP},target=${RPCP},mode=ingress \
    --replicas 1 \
    --with-registry-auth \
    --detach=false \
   ${REGISTRY_HOST}/${IMAGE} --nousb \
    --fakenet=$i/$N \
    --rpc --rpcaddr 0.0.0.0 --rpcport ${RPCP} --rpccorsdomain "*" --rpcapi "eth,debug,admin,web3" \
    --port ${PORT} --nat extip:${SWARM_HOST}

done