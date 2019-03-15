#!/usr/bin/env bash

set -euxo pipefail


# Constants

declare -ri N=${1:-4}
declare -r MPWD="$(pwd)"

declare -r network_name='lachesisnet'


# Cleanup

## Containers
containers="$(docker ps -f name=lachesis_ -aq)"
if [ ! -z "$containers" ]; then
    docker stop $containers
    docker rm $containers
fi

## Network
if [[ $(docker network ls -qf name="$network_name") != "" ]]; then
    docker network rm "$network_name"
fi

## Keys

if [ -d conf ]; then
    rm -r conf
fi


## Creation

## Keys
batch-ethkey -network 172.77.5.0 -dir conf -n "$N"

## Network
docker network create \
  --driver=bridge \
  --subnet=172.77.0.0/16 \
  --ip-range=172.77.5.0/24 \
  --gateway=172.77.5.254 \
  "$network_name"

## Dummy

### Docker

#### Image
if [ ! $(docker images go-dummy -q) ]; then
  pushd ..
  make build
  popd
fi

#### Containers
for ((i=1;i<=N;i++));
do
    docker run -d --name="lachesis_client$i" --net="$network_name" --ip=172.77.5.$(($N+$i)) -it go-dummy \
    --name="lachesis_client $i" \
    --client-listen="172.77.5.$(($N+$i)):1339" \
    --proxy-connect="172.77.5.$i:1338" \
    --discard \
    --log="debug" 
done

## Go-lachesis
for ((i=1;i<=N;i++));
do
    name="lachesis_node$(($i-1))"
    docker create --name="$name" --net="$network_name" --ip="172.77.5.$i" go-lachesis run \
    --cache-size=50000 \
    --timeout=200ms \
    --heartbeat=10ms \
    --listen="172.77.5.$i:1337" \
    --proxy-listen="172.77.5.$i:1338" \
    --client-connect="172.77.5.$(($N+$i)):1339" \
    --service-listen="172.77.5.$i:80" \
    --sync-limit=1000 \
    --store \
    --log="debug"

    docker cp "$MPWD/conf/"$(($i-1)) "$name":/.lachesis
    docker start "$name"
done
