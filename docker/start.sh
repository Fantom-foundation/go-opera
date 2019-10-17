#!/usr/bin/env bash
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
	--rpc --rpcapi "eth,debug,admin,web3" --nousb --verbosity 3 \
	${SENTRY_DSN}
done

attach_and_exec() {
    local NAME=$1
    local CMD=$2

    for attempt in $(seq 10)
    do
        echo -n "  - attempt ${attempt}: " >&2
        res=$(docker exec -i ${NAME} /lachesis --exec "${CMD}" attach http://127.0.0.1:18545)
        if [ $? -eq 0 ]
        then
            echo " ok" >&2
            echo $res
            break
        else
            sleep 1
        fi
    done
}

echo -e "\nConnect nodes to ring:\n"
for i in $(seq $N)
do
    j=$((i % N + 1))

    echo " getting node-$j address:"
    ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${NAME}-$j)
    enode=$(attach_and_exec ${NAME}-$j 'admin.nodeInfo.enode')
    enode=$(echo ${enode} | sed "s/127.0.0.1/${ip}/")
    echo "    address = ${enode}"

    echo " connecting node-$i to node-$j:"
    res=$(attach_and_exec ${NAME}-$i "admin.addPeer(${enode})")
    echo "    result = ${res}"
done


. ./_prometheus.sh
