lachesis builder
================

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
    docker stop $(docker ps -a --no-trunc --filter name='^/lachesis' --format '{{.Names}}')
