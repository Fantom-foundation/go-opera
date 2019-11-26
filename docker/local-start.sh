#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

set -e


echo -e "\nStart $N nodes:\n"

for ((i=$N-1;i>=0;i-=1))
do
    rm -f ./transactions.rlp
    DATADIR="${PWD}/.lachesis$i"
    rm -fr ${DATADIR}
    mkdir -p ${DATADIR}

    PORT=$(($PORT_BASE+$i))
    RPCP=$(($RPCP_BASE+$i))
    WSP=$(($WSP_BASE+$i))
    ACC=$(($i+1)) 
    (go run ../cmd/lachesis \
	--datadir=${DATADIR} \
	--fakenet=${ACC}/$N \
	--port=${PORT} \
	--rpc --rpcaddr="127.0.0.1" --rpcport=${RPCP} --rpccorsdomain="*" --rpcapi="eth,debug,admin,web3" \
	--ws --wsaddr="127.0.0.1" --wsport=${WSP} --wsorigins="*" --wsapi="eth,debug,admin,web3,personal" \
	--nousb --verbosity=5 --tracing  &> .lachesis$i.log)&
done

attach_and_exec() {
    local i=$1
    local CMD=$2
    local RPCP=$(($RPCP_BASE+$i))

    for attempt in $(seq 20)
    do
        if (( attempt > 5 ));
        then 
            echo "  - attempt ${attempt}: " >&2
        fi;

        res=$(go run ../cmd/lachesis --exec "${CMD}" attach http://127.0.0.1:${RPCP} 2> /dev/null)
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
for ((i=$N-1;i>=0;i-=1))
do
    j=$(((i+1) % N))

    enode=$(attach_and_exec $j 'admin.nodeInfo.enode')
    echo "    p2p address = ${enode}"

    echo " connecting node-$i to node-$j:"
    res=$(attach_and_exec $i "admin.addPeer(${enode})")
    echo "    result = ${res}"
done
