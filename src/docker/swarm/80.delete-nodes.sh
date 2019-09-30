#!/bin/bash

source $(dirname $0)/set_env.sh


for i in `seq 1 $N`
do

  NAME=node$i

  docker $SWARM service rm \
    ${NAME}

done