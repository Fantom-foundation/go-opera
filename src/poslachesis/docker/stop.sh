#!/usr/bin/env bash
cd $(dirname $0)

docker ps -q --filter "network=lachesis" | while read id
do
    docker stop $id
done

sentry/stop.sh
