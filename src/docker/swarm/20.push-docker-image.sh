#!/bin/bash

source $(dirname $0)/set_env.sh


IMAGE=lachesis:latest

docker tag ${IMAGE} ${REGISTRY_HOST}/${IMAGE}
docker push ${REGISTRY_HOST}/${IMAGE}
