#!/bin/bash

source $(dirname $0)/set_env.sh


sudo mkdir -p /etc/docker/certs.d/${REGISTRY_HOST}
sudo cp ${CA_FILE} /etc/docker/certs.d/${REGISTRY_HOST}/ca.crt
sudo systemctl restart docker

docker login ${REGISTRY_HOST}