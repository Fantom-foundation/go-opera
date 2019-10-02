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
    sleep 2
done

echo -e "\nConnect nodes (ring):\n"
for i in $(seq $N)
do
    sleep 2
    j=$((i % N + 1))

    ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${NAME}-$j)

    enode=$(docker exec -i ${NAME}-$j /lachesis --exec 'admin.nodeInfo.enode' attach http://127.0.0.1:18545)
    enode=$(echo $enode | sed "s/127.0.0.1/${ip}/")

    docker exec -i ${NAME}-$i /lachesis --exec "admin.addPeer(${enode})" attach http://127.0.0.1:18545
    echo "Connected $i to $j, with $enode"
done


. ./_prometheus.sh
