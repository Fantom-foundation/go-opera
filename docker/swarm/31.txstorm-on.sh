#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


# temporary solution, revers because testnet0 is not avaible from Russia
for ((i=$N-1;i>=0;i-=1))
do

  NAME=txgen$i
  RPCP=$(($RPCP_BASE+$i))
  PART=$(($i+1))

  docker $SWARM service create \
    --network lachesis \
    --hostname="{{.Service.Name}}" \
    --name ${NAME} \
    --replicas 1 \
    --with-registry-auth \
    --detach=false \
   ${REGISTRY_HOST}/tx-storm:${TAG} \
    --num=$PART/$N --rate=10 --period=30 \
    --verbosity 5 --metrics \
    http://node$i:${RPCP} 

done
