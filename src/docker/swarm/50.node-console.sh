#!/bin/bash

source $(dirname $0)/set_env.sh

i=$1
shift

# temporary solution, remove when ingress on                                                                                                                                                              
HOST=testnet$(($i-1))                                                                                                                                                                                     
SWARM_HOST=`./swarm node inspect $HOST --format "{{.Status.Addr}}"`

RPCP=$(($RPCP_BASE+$i))

docker run --rm -ti ${IMAGE} $@ attach http://${SWARM_HOST}:${RPCP}

