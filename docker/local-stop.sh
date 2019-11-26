#!/usr/bin/env bash
cd $(dirname $0)

killall tx-storm
killall lachesis
docker stop tracing