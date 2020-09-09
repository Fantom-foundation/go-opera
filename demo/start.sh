#!/bin/bash

# This script will launch a cluster of Lachesis nodes
# The parameter N = number of nodes to run

set -e

# number of nodes N
N=5

#
PROG=opera
EXEC=../build/opera

# default ip using localhost
IP=127.0.0.1
# default RPC port PORT
# the actual ports are PORT+1, PORT+2, etc (18541, 18542, 18543, ... )
#PORT=18540
PORT=4000

declare -r LACHESIS_BASE_DIR=/tmp/opera-demo


echo -e "\nStart $N nodes:"
for i in $(seq $N)
do
    rpcport=$((PORT + i))
    p2pport=$((5050 + i))

    ${EXEC} \
  --nat extip:$IP \
	--fakenet $i/$N \
	--port ${p2pport} --rpc --rpcapi "eth,ftm,dag,debug,admin,web3,personal,net,txpool" --rpcport ${rpcport} --nousb --verbosity 3 \
	--datadir "${LACHESIS_BASE_DIR}/datadir/opera$i" &
    echo -e "Started opera client at ${IP}:${port}"
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
