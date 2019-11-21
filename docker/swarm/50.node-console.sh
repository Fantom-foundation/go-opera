#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


i=$1
shift

# temporary solution, remove when ingress on
HOST=testnet$i
SWARM_HOST=`./swarm node inspect $HOST --format "{{.Status.Addr}}"`

RPCP=$(($RPCP_BASE+$i))

docker run --rm -i lachesis:${TAG} $@ attach http://${SWARM_HOST}:${RPCP}

