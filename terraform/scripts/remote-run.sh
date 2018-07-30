#!/bin/bash

set -eux
echo $0

private_ip=${1}
public_ip=${2}

ssh -q -i lachesis.pem -o "UserKnownHostsFile /dev/null" -o "StrictHostKeyChecking=no" \
 ubuntu@$public_ip  <<-EOF
    nohup /home/ubuntu/bin/lachesis run \
    --datadir=/home/ubuntu/lachesis_conf \
    --store=inmem \
    --cache_size=10000 \
    --tcp_timeout=500 \
    --heartbeat=50 \
    --node_addr=$private_ip:1337 \
    --proxy_addr=:1338 \
    --client_addr=:1339 \
    --service_addr=:8080 > lachesis_logs 2>&1 &

    nohup /home/ubuntu/bin/dummy \
    --name=$public_ip \
    --client_addr=:1339 \
    --proxy_addr=:1338 < /dev/null > dummy_logs 2>&1 &
EOF
