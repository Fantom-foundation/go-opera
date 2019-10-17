#!/usr/bin/env bash

set -e

cd $(dirname $0)

. ./_params.sh

NETWORK=lachesis
docker network create ${NETWORK}

. ./_sentry.sh

echo -e "\nStart $N nodes:\n"
for i in $(seq $N)
do
    docker run -d --rm \
	--net=${NETWORK} --name=${NAME}-$i \
	--cpus=${LIMIT_CPU} --blkio-weight=${LIMIT_IO} \
	"lachesis" \
	--fakenet $i/$N \
	--port 5050 --rpc --rpcapi "eth,debug,admin,web3" --rpcport 18545 --nousb --verbosity 3 \
	${SENTRY_DSN}
done

attach_and_exec() {
    local NAME=$1
    local CMD=$2

    for attempt in $(seq 20)
    do
        if (( attempt > 5 ));
        then 
            echo "  - attempt ${attempt}: " >&2
        fi;

        res=$(docker exec -i ${NAME} /lachesis --exec "${CMD}" attach http://127.0.0.1:18545 2> /dev/null)
        if [ $? -eq 0 ]
        then
            #echo "success" >&2
            echo $res
            return 0
        else
            #echo "wait" >&2
            sleep 1
        fi
    done
    echo "failed RPC connection to ${NAME}" >&2
    return 1
}

echo -e "\nConnect nodes to ring:\n"
for i in $(seq $N)
do
    j=$((i % N + 1))

    echo " getting node-$j address:"
    ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${NAME}-$j)
    enode=$(attach_and_exec ${NAME}-$j 'admin.nodeInfo.enode')
    enode=$(echo ${enode} | sed "s/127.0.0.1/${ip}/")
    echo "    p2p address = ${enode}"

    echo " connecting node-$i to node-$j:"
    res=$(attach_and_exec ${NAME}-$i "admin.addPeer(${enode})")
    echo "    result = ${res}"
done


. ./_prometheus.sh
