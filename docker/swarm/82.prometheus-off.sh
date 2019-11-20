#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


docker $SWARM service rm prometheus
docker $SWARM config rm prometheus
