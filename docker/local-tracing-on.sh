#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

set -e

docker run -d --rm \
    --name tracing \
    -p 5775:5775/udp \
    -p 6831:6831/udp \
    -p 6832:6832/udp \
    -p 5778:5778 \
    -p 14268:14268 \
    -p 14250:14250 \
    -p 16686:16686 \
    jaegertracing/all-in-one
