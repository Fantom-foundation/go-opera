#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


i=$1
shift

RPCP=$(($RPCP_BASE+$i))

docker run --rm -i lachesis:${TAG} $@ attach http://${SWARM_HOST}:${RPCP}

