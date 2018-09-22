#!/usr/bin/env bash

docker network rm lachesis-net
docker network create --driver=bridge --subnet="$1" lachesis-net

