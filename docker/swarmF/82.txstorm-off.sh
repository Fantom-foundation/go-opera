#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


for ((i=$N-1;i>=0;i-=1))
do

  NAME=txgen$i

  docker $SWARM service rm \
    ${NAME}

done