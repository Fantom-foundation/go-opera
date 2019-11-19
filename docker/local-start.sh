#!/usr/bin/env bash
cd $(dirname $0)

set -e

. ./_params.sh

echo -e "\nStart $N nodes:\n"

for i in $(seq $N)
do
    rm -f ./transactions.rlp
    DATADIR="${PWD}/.lachesis$i"
    rm -fr ${DATADIR}
    mkdir -p ${DATADIR}
    (go run ../cmd/lachesis \
	--datadir=${DATADIR} \
	--fakenet $i/$N \
	--port $((5050+i)) --rpc --rpcaddr 127.0.0.1 --rpcport $((4000+i)) --rpccorsdomain "*" --rpcapi "eth,debug,admin,web3" \
	--nousb --verbosity 5  &> .lachesis$i.log)&
done

attach_and_exec() {
    local i=$1
    local CMD=$2

    for attempt in $(seq 20)
    do
        if (( attempt > 5 ));
        then 
            echo "  - attempt ${attempt}: " >&2
        fi;

        res=$(go run ../cmd/lachesis --exec "${CMD}" attach http://127.0.0.1:$((4000+i)) 2> /dev/null)
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

    enode=$(attach_and_exec $j 'admin.nodeInfo.enode')
    echo "    p2p address = ${enode}"

    echo " connecting node-$i to node-$j:"
    res=$(attach_and_exec $i "admin.addPeer(${enode})")
    echo "    result = ${res}"
done
