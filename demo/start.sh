#!/usr/bin/env bash


cd $(dirname $0)
. ./_params.sh
if test -f "./_localparams.sh"; then
  . ./_localparams.sh
fi

set -e

echo -e "\nStart $N nodes:\n"

go build -o ../build/demo_opera ../cmd/opera

rm -f ./transactions.rlp
for ((i=0;i<$N;i+=1))
do
    DATADIR="${PWD}/opera$i.datadir"
    mkdir -p ${DATADIR}

    PORT=$(($PORT_BASE+$i))
    RPCP=$(($RPCP_BASE+$i))
    WSP=$(($WSP_BASE+$i))
    ACC=$(($i+1))
    (../build/demo_opera \
	--datadir=${DATADIR} \
	--fakenet=${ACC}/$N \
	--port=${PORT} \
	--nat extip:127.0.0.1 \
	--http --http.addr="127.0.0.1" --http.port=${RPCP} --http.corsdomain="*" --http.api="eth,debug,net,admin,web3,personal,txpool,ftm,dag" \
	--ws --ws.addr="127.0.0.1" --ws.port=${WSP} --ws.origins="*" --ws.api="eth,debug,net,admin,web3,personal,txpool,ftm,dag" \
	--metrics --metrics.addr=127.0.0.1 --metrics.port=$(($RPCP+1100)) $METRICS_ARGS
	--verbosity=3 --tracing >> opera$i.log 2>&1)&

    echo -e "\tnode$i ok"
done

echo -e "\nConnect nodes to ring:\n"
for ((i=0;i<$N;i+=1))
do
    for ((n=0;n<$M;n+=1))
    do
        j=$(((i+n+1) % N))

	enode=$(attach_and_exec $j 'admin.nodeInfo.enode')
        echo "    p2p address = ${enode}"

        echo " connecting node-$i to node-$j:"
        res=$(attach_and_exec $i "admin.addPeer(${enode})")
        echo "    result = ${res}"
    done
done
