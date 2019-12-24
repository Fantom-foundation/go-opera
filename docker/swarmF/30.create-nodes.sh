#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


bootnode=""
#bootnode="--bootnodes enode://de84c5886b42a022fdca1bc99ac5c464ac71bbb10948c8131efec30d4c0fb2da5c6082692c60597fcad97b8858d44cd048939fdd1dc217e00140a2729112ae07@18.191.96.173:3002"

# temporary solution, revers because testnet0 is not avaible from Russia
for ((i=$N-1;i>=0;i-=1))
#for i in 3
do
  NAME=node$i
  PORT=$(($PORT_BASE+$i))
  RPCP=$(($RPCP_BASE+$i))
  WSP=$(($WSP_BASE+$i))
  ACC=$(($i+1))

  # temporary solution, remove when ingress on
  HOST=testnet$i
  SWARM_HOST=`./swarm node inspect $HOST --format "{{.Status.Addr}}"`

  docker $SWARM service inspect ${NAME} &>/dev/null || \
  docker $SWARM service create \
    --name ${NAME} \
    --publish published=${PORT},target=${PORT},mode=ingress,protocol=tcp \
    --publish published=${PORT},target=${PORT},mode=ingress,protocol=udp \
    --publish published=${RPCP},target=${RPCP},mode=ingress \
    --publish published=${WSP},target=${WSP},mode=ingress \
    --mount type=bind,source=/var/lachesis,destination=/var/lachesis \
    --replicas 1 \
    --with-registry-auth \
    --detach=false \
    --constraint node.hostname==$HOST \
   ${REGISTRY_HOST}/lachesis:${TAG} --nousb \
    --verbosity=3 \
    --fakenet=$ACC/$N \
    --rpc --rpcaddr="0.0.0.0" --rpcport=${RPCP} --rpccorsdomain="*" --rpcapi="eth,debug,admin,web3,personal,net,txpool,ftm,sfc" \
    --ws --wsaddr="0.0.0.0" --wsport=${WSP} --wsorigins="*" --wsapi="eth,debug,admin,web3,personal,net,txpool,ftm,sfc" \
    --port=${PORT} --nat="extip:${SWARM_HOST}" \
    --datadir=/var/lachesis \
    ${bootnode}

    if [ -z "$bootnode" ]
    then
        sleep 6
        enode=`./50.node-console.sh $i --exec 'admin.nodeInfo.enode' | xargs`
        echo "Enode of ${NAME} is ${enode}"
        bootnode="--bootnodes=${enode}"
    fi

done
