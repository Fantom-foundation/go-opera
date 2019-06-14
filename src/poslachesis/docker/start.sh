#!/usr/bin/env bash
cd $(dirname $0)

. ./_params.sh

NETWORK=lachesis
docker network create ${NETWORK}

. ./_sentry.sh

echo -e "\nStart $N nodes:\n"
for i in $(seq $N)
do
    j=$((i % N + 1)) # ring
    docker run -d --rm \
	--net=${NETWORK} --name=${NAME}-$i \
	--cpus=${LIMIT_CPU} --blkio-weight=${LIMIT_IO} \
	"pos-lachesis" \
	start --network=fake:$i/$N --peer=${NAME}-$j --db=/tmp ${SENTRY_DSN} --metrics
done

. ./_prometheus.sh
