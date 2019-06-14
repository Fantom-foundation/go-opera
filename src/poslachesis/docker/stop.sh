#!/usr/bin/env bash
cd $(dirname $0)

. ./_params.sh

NETWORK=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}' ${NAME}-1)


NETWORK=${NETWORK} sentry/stop.sh

docker ps -q --filter "network=${NETWORK}" | while read id
do
    docker stop $id
done

blockade destroy

docker network rm ${NETWORK}
