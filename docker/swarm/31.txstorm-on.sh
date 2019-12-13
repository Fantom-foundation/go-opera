#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


for ((i=0;i<$N;i+=1))
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
    --num=$PART/$N --rate=10 \
    --accs-start=${TEST_ACCS_START} --accs-count=${TEST_ACCS_COUNT} \
    --verbosity 3 --metrics \
    http://node$i:${RPCP} 

done
