#!/bin/bash

source $(dirname $0)/set_env.sh


bootnode=""

# temporary solution, revers because testnet0 is not avaible from Russia
for ((i=$N-1;i>=0;i-=1))
do

  NAME=node$i
  PORT=$(($PORT_BASE+$i))
  RPCP=$(($RPCP_BASE+$i))
  ACC=$(($i+1))

  # temporary solution, remove when ingress on
  HOST=testnet$i
  SWARM_HOST=`./swarm node inspect $HOST --format "{{.Status.Addr}}"`

  docker $SWARM service create \
    --name ${NAME} \
    --publish published=${PORT},target=${PORT},mode=ingress,protocol=tcp \
    --publish published=${PORT},target=${PORT},mode=ingress,protocol=udp \
    --publish published=${RPCP},target=${RPCP},mode=ingress \
    --replicas 1 \
    --with-registry-auth \
    --detach=false \
    --constraint node.hostname==$HOST \
   ${REGISTRY_HOST}/${IMAGE} --nousb \
    --fakenet=$ACC/$N \
    --rpc --rpcaddr 0.0.0.0 --rpcport ${RPCP} --rpccorsdomain "*" --rpcapi "eth,debug,admin,web3" \
    --port ${PORT} --nat extip:${SWARM_HOST} \
    ${bootnode}

    if [ -z "$bootnode" ]
    then
        sleep 6
        enode=`./50.node-console.sh $i --exec 'admin.nodeInfo.enode' | xargs`
        echo "Enode of ${NAME} is ${enode}"
        bootnode="--bootnodes ${enode}"
    fi

done