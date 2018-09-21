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

    rm -rf nodes peers.json
    
    nodes_num=1000
    ip_start='192.168.0.0'
    cidr='/16'
    ip_range="$ip/$cidr"
    
    batch-ethkey -dir nodes -network "$ip_start" -n 1000 > peers.json
    ./network.bash "$ip_range"
    ./spin_multi.bash 1000 "$ip_range"
