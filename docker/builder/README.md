lachesis builder
================

## Generate private & public keys, and peers.json

    "$GOPATH/src/github.com/andrecronje/evm/scripts/gen_peers.sh" nodes 3 > peers.json

## SSL certs

    cp -r /etc/ssl/certs certs

## Docker build command

    docker build --compress --squash --force-rm --tag lachesis0 --build-arg ca_certificates=certs .

## Docker run command

    docker run --rm lachesis0
