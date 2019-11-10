#!/bin/bash

source $(dirname $0)/set_env.sh


docker $SWARM service rm prometheus
docker $SWARM config rm prometheus
