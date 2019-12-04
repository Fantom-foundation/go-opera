#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


sudo mkdir -p /etc/docker/certs.d/${REGISTRY_HOST}
sudo cp ${CA_FILE} /etc/docker/certs.d/${REGISTRY_HOST}/ca.crt
sudo systemctl restart docker

docker login ${REGISTRY_HOST}