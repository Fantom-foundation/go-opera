lachesis builder
================

## Generate private & public keys, and peers.json

    go get -v github.com/SamuelMarks/batch-ethkey

    batch-ethkey -dir nodes -n 3 > peers.json

## SSL certs

    cp -r /etc/ssl/certs certs

## Docker build command

    docker build --compress --squash --force-rm --tag 'lachesis' .

## Docker run command

    node_num=0; docker run -e node_num="$node_num" -p $(( 12000+"$node_num" )):1339 --name "lachesis$node_num" --rm 'lachesis'

Or just use the script:

    ./spin.bash 0
