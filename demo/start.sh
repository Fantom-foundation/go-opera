#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

set -e

TOTAL=$((N+M))
echo -e "\nStart ${TOTAL} validators on $N-validator genesis:\n"

go build -o ../build/demo_opera ../cmd/opera
rm -f ./transactions.rlp

for ((i=0;i<${TOTAL};i+=1))
do
    run_opera_node $i &
    echo -e "\tnode$i ok"
done

echo -e "\nConnect nodes to ring:\n"
for ((i=0;i<${TOTAL};i+=1))
do
	j=$(((i+n+1) % TOTAL))

	enode=$(attach_and_exec $j 'admin.nodeInfo.enode')
	echo "    p2p address = ${enode}"

	echo " connecting node-$i to node-$j:"
	res=$(attach_and_exec $i "admin.addPeer(${enode})")
	echo "    result = ${res}"
done
