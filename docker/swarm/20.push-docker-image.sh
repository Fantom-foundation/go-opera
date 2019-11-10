#!/bin/bash

source $(dirname $0)/set_env.sh

for IMAGE in ${NODE_IMAGE} ${TXSTORM_IMAGE}
do
    docker tag ${IMAGE} ${REGISTRY_HOST}/${IMAGE}
    docker push ${REGISTRY_HOST}/${IMAGE}
done
