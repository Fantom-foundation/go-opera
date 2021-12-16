#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

set -e

echo -e "\nStart node 9 with snapsync:\n"

go build -o ../build/opera_n9 ../cmd/opera

i=9
    DATADIR="${PWD}/opera$i.datadir"
    mkdir -p ${DATADIR}

    PORT=$(($PORT_BASE+$i))
    RPCP=$(($RPCP_BASE+$i))
    WSP=$(($WSP_BASE+$i))
    (../build/opera_n9 \
	--datadir=${DATADIR} \
	--fakenet=0/$N \
	--port=${PORT} \
	--nat extip:127.0.0.1 \
	--http --http.addr="127.0.0.1" --http.port=${RPCP} --http.corsdomain="*" --http.api="eth,debug,net,admin,web3,personal,txpool,ftm,dag" \
	--ws --ws.addr="127.0.0.1" --ws.port=${WSP} --ws.origins="*" --ws.api="eth,debug,net,admin,web3,personal,txpool,ftm,dag" \
	--gcmode full \
	--syncmode=snap \
	--verbosity=3 --tracing &>> opera$i.log)&

    echo -e "\tnode$i ok"

echo -e "\nConnect nodes to ring:\n"

i=9
    for n in 1 2
    do
        j=$(((i+n) % N))

	enode=$(attach_and_exec $j 'admin.nodeInfo.enode')
        echo "    p2p address = ${enode}"

        echo " connecting node-$i to node-$j:"
        res=$(attach_and_exec $i "admin.addPeer(${enode})")
        echo "    result = ${res}"
    done

