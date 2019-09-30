#!/bin/bash

source $(dirname $0)/set_env.sh


for i in `seq 1 $N`
do

  NAME=node$i

  docker $SWARM service update ${NAME} \
    --stop-grace-period 10s \
    --image ${REGISTRY_HOST}/${IMAGE} \
    --with-registry-auth \
    --detach=false

done