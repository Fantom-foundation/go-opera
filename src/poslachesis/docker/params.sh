#!/usr/bin/env bash

# Sentry
declare -rl SENTRY="${SENTRY:-no}"
if [ "$SENTRY" == "yes" ]; then
    sentry/start.sh && source sentry/.dsn
    DSN=--dsn="$dsn"
fi

# Lachesis
declare -ri N="${N:-3}"

limit_cpu=$(echo "scale=2; 1/$N" | bc)
limit_io=$(echo "500/$N" | bc)
limits="--cpus=${limit_cpu} --blkio-weight=${limit_io}"
