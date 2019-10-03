#!/bin/bash

source $(dirname $0)/set_env.sh


for ((i=$N-1;i>=0;i-=1))
do

  NAME=node$i

  docker $SWARM service rm \
    ${NAME}

done