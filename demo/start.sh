#!/bin/bash

# This script will launch a cluster of Lachesis nodes
# The parameter N = number of nodes to run

set -e

# number of nodes N
N=5
LIMIT_CPU=$(echo "scale=2; 1/$N" | bc)
LIMIT_IO=$(echo "500/$N" | bc)

#
PROG=lachesis
EXEC=../build/lachesis

# default ip using localhost
IP=127.0.0.1
# default port PORT
# the actual ports are PORT+1, PORT+2, etc (18541, 18542, 18543, ... )
#PORT=18540
PORT=4000

declare -r LACHESIS_BASE_DIR=/tmp/lachesis-demo


echo -e "\nStart $N nodes:"
for i in $(seq $N)
do
    port=$((PORT + i))
    localport=$((5050 + i))

    ${EXEC} \
	--fakenet $i/$N \
	--port ${localport} --rpc --rpcapi "eth,ftm,debug,admin,web3,personal,net,txpool" --rpcport ${port} --nousb --verbosity 3 \
	--datadir "${LACHESIS_BASE_DIR}/datadir/lach$i" &
    echo -e "Started lachesis client at ${IP}:${port}"
done



attach_and_exec() {
    local URL=$1
    local CMD=$2

    for attempt in $(seq 20)
    do
        if (( attempt > 5 ));
        then
            echo "  - attempt ${attempt}: " >&2
        fi;

        res=$("${EXEC}" --exec "${CMD}" attach http://${URL} 2> /dev/null)
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
	url=${IP}:$((PORT + j))
	echo "    at url: ${url}"

    enode=$(attach_and_exec ${url} 'admin.nodeInfo.enode')
    echo "    p2p address = ${enode}"

    echo " connecting node-$i to node-$j:"
    url=${IP}:$((PORT + i))
    echo "    at url: ${url}"

    res=$(attach_and_exec ${url} "admin.addPeer(${enode})")
    echo "    result = ${res}"
done


echo
echo "Sleep for 10000 seconds..."
sleep 10000
