#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

echo "Don't use it without persistent db volume, cause --bootnodes becomes invalid" >&2
#exit

#for ((i=$N-1;i>=0;i-=1))
#do
i=3
  NAME=node$i

  docker $SWARM service update ${NAME} \
    --stop-grace-period 10s \
    --image ${REGISTRY_HOST}/lachesis:${TAG} \
    --with-registry-auth \
    --detach=false

#done