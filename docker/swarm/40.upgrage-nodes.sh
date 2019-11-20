#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


for ((i=$N-1;i>=0;i-=1))
do

  NAME=node$i

  docker $SWARM service update ${NAME} \
    --stop-grace-period 10s \
    --image ${REGISTRY_HOST}/lachesis:${TAG} \
    --with-registry-auth \
    --detach=false

done