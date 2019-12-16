#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

docker $SWARM network inspect lachesis &>/dev/null || \
docker $SWARM network create --driver overlay lachesis


bootnode=""
for ((i=$N-1;i>=0;i-=1))
do
  NAME=node$i
  PORT=$(($PORT_BASE+$i))
  RPCP=$(($RPCP_BASE+$i))
  WSP=$(($WSP_BASE+$i))
  ACC=$(($i+1))

  # github.com/jaegertracing/jaeger-client-go#environment-variables

  docker $SWARM service inspect ${NAME} &>/dev/null || \
  docker $SWARM service create \
    --network lachesis \
    --hostname="{{.Service.Name}}" \
    --name ${NAME} \
    --publish ${PORT}:${PORT}/tcp \
    --publish ${PORT}:${PORT}/udp \
    --publish ${RPCP}:${RPCP} \
    --publish ${WSP}:${WSP} \
    --env JAEGER_AGENT_HOST=tracing \
    --env JAEGER_AGENT_PORT=6831 \
    --env JAEGER_SAMPLER_MANAGER_HOST_PORT=tracing:5778 \
    --replicas 1 \
    --with-registry-auth \
    --detach=false \
   ${REGISTRY_HOST}/lachesis:${TAG} --nousb \
    --fakenet=$ACC/$N,/tmp/test_accs.json \
    --port=${PORT} --nat="extip:${SWARM_HOST}" \
    --rpc --rpcaddr="0.0.0.0" --rpcport=${RPCP} --rpcvhosts="*" --rpccorsdomain="*" --rpcapi="eth,debug,admin,web3,personal,net,txpool,ftm,sfc" \
    --ws --wsaddr="0.0.0.0" --wsport=${WSP} --wsorigins="*" --wsapi="eth,debug,admin,web3,personal,net,txpool,ftm,sfc" \
    --verbosity=3 --metrics --tracing \
    ${bootnode}

    if [ -z "$bootnode" ]
    then
        sleep 6
        enode=`./50.node-console.sh $i --exec 'admin.nodeInfo.enode' | xargs`
        echo "Enode of ${NAME} is ${enode}"
        bootnode="--bootnodes=${enode}"
    fi

done
