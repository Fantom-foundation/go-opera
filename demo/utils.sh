#!/bin/bash

declare -a pids

# number of nodes N
N=3
# number of new nodes M
M=3
# total
T=$((N+M))

set -e
LIMIT_CPU=$(echo "scale=2; 1/$N" | bc)
LIMIT_IO=$(echo "500/$N" | bc)

# base dir for running demo
LACHESIS_BASE_DIR=/tmp/lachesis-demo

#
PROG=lachesis
EXEC=../build/lachesis

# default ip using localhost
IP=127.0.0.1
# default port PORT
# the actual ports are PORT+1, PORT+2, etc (18641, 18642, 18643, ... )
PORT=18800
# default base local port
# actual local ports are LOCALPORT+1, LOCALPORT+2, ...
LOCALPORT=5800


# Function to be called when user press ctrl-c.
ctrl_c() {
    echo "Shell terminating ..."
    for pid in "${pids[@]}"
    do
	echo "Killing ${pid}..."
	kill -9 ${pid}
	# Suppresses "[pid] Terminated ..." message.
	wait ${pid} &>/dev/null
    done
    exit;
}

start_node() {
	local i=$1
	local N=$2
	echo -e "start_node node $i:"

    port=$((PORT + i))
	localport=$((LOCALPORT + i))

	echo -e "port=${port}, localport=${localport} "

    ${EXEC} \
	--fakenet $i/$N \
	--port ${localport} --rpc --rpcapi "eth,dag,debug,admin,web3" --rpcport ${port} --nousb --verbosity 3 \
	--datadir "${LACHESIS_BASE_DIR}/datadir/lach$i" &
	pids+=($!)
    echo -e "Started lachesis client at ${IP}:${port}, pid: $!"
    echo -e "\n"
}

start_nodes() {
	local i=$1
	local j=$2
	local N=$3
	echo -e "\nStart $N nodes:\n"
	for i in $(seq $i $j)
	do
	    start_node $i $N
	done
}

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

connect_pair() {
    local from=$1
    local to=$2

	echo " getting node-${to} address:"
	url=${IP}:$((PORT + to))
	echo "    at url: ${url}"

    enode=$(attach_and_exec ${url} 'admin.nodeInfo.enode')
    echo "    p2p address = ${enode}"

    echo " connecting node-${from} to node-${to}:"
    url=${IP}:$((PORT + from))
    echo "    at url: ${url}"

    res=$(attach_and_exec ${url} "admin.addPeer(${enode})")
    echo "    result = ${res}"

	return 0
}

connect_nodes() {
	local from=$1
	local to=$2

	echo -e "\nConnect nodes to ring:\n"
	for i in $(seq ${from} $((to-1)))
	do
	    j=$((i + 1))
	    conn=$(connect_pair $i $j)
	done
	conn=$(connect_pair ${to} ${from})
}


