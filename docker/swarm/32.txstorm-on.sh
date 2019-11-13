#!/bin/bash

source $(dirname $0)/set_env.sh


# temporary solution, revers because testnet0 is not avaible from Russia
for ((i=$N-1;i>=0;i-=1))
do

  NAME=txstorm$i
  PORT=$(($PORT_BASE+$i))
  RPCP=$(($RPCP_BASE+$i))
  ACC=$(($i+1))

  # temporary solution, remove when ingress on
  HOST=testnet$i
  SWARM_HOST=`./swarm node inspect $HOST --format "{{.Status.Addr}}"`

  docker $SWARM service create \
    --name ${NAME} \
    --replicas 1 \
    --with-registry-auth \
    --detach=false \
    --constraint node.hostname==$HOST \
   ${REGISTRY_HOST}/${TXSTORM_IMAGE} \
    --num=$ACC/$N --donor=$ACC --rate=500 --period=120 --metrics \
    http://${SWARM_HOST}:${RPCP} 

done
