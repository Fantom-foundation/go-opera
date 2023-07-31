#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

set -e

TOTAL=$((N+M))
i=$1
DATADIR="$(data_dir $i)"
if [ -d ${DATADIR} ]; then
	echo "${DATADIR} already exists!"
	exit 1
fi
mkdir -p "${DATADIR}"

echo -e "\nStart additional node $i:\n"
(run_opera_node $i)&
echo -e "\tnode$i ok"

echo -e "\nConnect to existing nodes:\n"
for ((n=0;n<$TOTAL;n+=1))
do
	enode=$(attach_and_exec $n 'admin.nodeInfo.enode')
	echo "    p2p address = ${enode}"

	echo " connecting node-$i to node-$n:"
	res=$(attach_and_exec $i "admin.addPeer(${enode})")
	echo "    result = ${res}"
done
sleep 5

VPKEY=$(grep "Unlocked validator key" opera$i.log | head -n 1 | sed 's/.* pubkey=\(0x.*\)$/\1/')
VADDR=$(grep "Unlocked fake validator account" opera$i.log | head -n 1 | sed 's/.* address=\(0x.*\)$/\1/')

echo -e "\nFund new validator acc ${VADDR}:\n"
attach_and_exec 0 "ftm.sendTransaction({from: personal.listAccounts[0], to: \"${VADDR}\", value: web3.toWei(\"510000.0\", \"ftm\")})"
sleep 5

echo -e "\nCall SFC to create validator ${VPKEY}:\n"
../build/demo_opera attach ./opera$i.datadir/opera.ipc << JS
	abi = JSON.parse('[{"constant":false,"inputs":[{"internalType":"bytes","name":"pubkey","type":"bytes"}],"name":"createValidator","outputs":[],"payable":true,"stateMutability":"payable","type":"function"}]');
	sfcc = web3.ftm.contract(abi).at("0xfc00face00000000000000000000000000000000");
	sfcc.createValidator("${VPKEY}", {from:"${VADDR}", value: web3.toWei("500000.0", "ftm")});
JS

echo -e "\nRestart the node:\n"
kill %1
sleep 10
(run_opera_node $i)&
