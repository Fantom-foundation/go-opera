#!/bin/bash

# This script will launch a cluster of N Lachesis nodes
# using 
# The parameter N = number of nodes to run

set -e

# number of nodes N
N=7
LIMIT_CPU=$(echo "scale=2; 1/$N" | bc)
LIMIT_IO=$(echo "500/$N" | bc)


FVM=${GOPATH}/src/github.com/Fantom-foundation/go-ethereum
BOOTNODE=${FVM}/build/bin/bootnode

echo "Generate bootnode.key"
${BOOTNODE} -genkey bootnode.key

echo "Start bootnode with bootnode.key"
bootnode=$( "${BOOTNODE}" -nodekey bootnode.key 2>/dev/null | head -1 & )
#bootnode=$( cat bootenode.txt )

echo -e "Bootnode=${bootnode}"

######
EXEC=../../build/lachesis

# default ip using localhost
IP=127.0.0.1
# default port PORT
# the actual ports are PORT+1, PORT+2, etc (18541, 18542, 18543, ... )
PORT=18540

# demo directory 
LACHESIS_BASE_DIR=/tmp/lachesis-demo

echo -e "\nStart $N nodes:"
for i in $(seq $N)
do
    port=$((PORT + i))
    localport=$((5050 + i))

    ${EXEC} \
	--bootnodes "${bootnode}" \
	--fakenet $i/$N \
	--port ${localport} --rpc --rpcapi "eth,ftm,debug,admin,web3" --rpcport ${port} --nousb --verbosity 3 \
	--datadir "${LACHESIS_BASE_DIR}/datadir/lach$i" &
    echo -e "Started lachesis client at ${IP}:${port}, pid: $!"
done



echo
echo "Sleep for 10000 seconds..."
sleep 10000
