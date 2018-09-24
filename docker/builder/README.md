lachesis builder
================

## Dependencies

  - [Docker](https://www.docker.com/get-started)
  - [jq](https://stedolan.github.io/jq)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey)
  - [upx](https://upx.github.io), download [upx-3.95-amd64_linux.tar.xz](https://github.com/upx/upx/releases/download/v3.95/upx-3.95-amd64_linux.tar.xz) and extract `upx` binary into this dir

## Generate private & public keys, and peers.json

    go get -v github.com/SamuelMarks/batch-ethkey

    batch-ethkey -dir nodes -network 192.168.0.0 -n 3 > peers.json

## SSL certs

    cp -r /etc/ssl/certs certs

## Docker build command

    docker build --compress --squash --force-rm --tag lachesis .

## Docker run command

Just use the script, with last arg specifying node number:

    ./spin.bash 0

## Larger scale testing

    n=10 ./scale.bash

## Cleanup

    rm -rf nodes peers.json
    containers="$(docker ps -a --no-trunc --filter name='^/lachesis' --format '{{.Names}}')"
    docker stop "$containers"
    docker rm  "$containers"
