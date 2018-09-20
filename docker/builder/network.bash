#!/usr/bin/env bash

docker network rm lachesis-net
docker network create --driver=bridge --subnet=192.168.0.0/16 lachesis-net

