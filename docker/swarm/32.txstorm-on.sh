#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


# temporary solution, revers because testnet0 is not avaible from Russia
for ((i=$N-1;i>=0;i-=1))
do

  NAME=txgen$i
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
   ${REGISTRY_HOST}/tx-storm:${TAG} \
    --num=${ACC}/$N --rate=10000 --period=30 \
    --verbosity 5 --metrics \
    http://${SWARM_HOST}:${RPCP} 

done
