#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


docker ps -a -q -f "network=${NETWORK}" -f "name=prometheus" | while read id
do
    docker stop $id 2>/dev/null # fine if stopped already
    docker rm $id 2>/dev/null # fine if removed already
done
