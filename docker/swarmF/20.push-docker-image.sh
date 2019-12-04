#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


for IMAGE in lachesis tx-storm
do
    docker tag ${IMAGE}:${TAG} ${REGISTRY_HOST}/${IMAGE}:${TAG}
    docker push ${REGISTRY_HOST}/${IMAGE}:${TAG}
done
