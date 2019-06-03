#!/usr/bin/env bash

docker ps -q --filter "network=lachesis" | while read id
do
    docker stop $id
done

docker network rm lachesis

# Close sentry
cd sentry && docker-compose down && cd ..
