#!/usr/bin/env bash
cd $(dirname $0)

CONF=prometheus.yml

cat << HEADER > $CONF
scrape_configs:

  - job_name: 'prometheus'
    static_configs:
      - targets: ['127.0.0.1:9090']

HEADER

for i in $(seq $N)
do

    ip=$(docker inspect $name-$i | grep IPAddress | tail -n1)
    ip=${ip//\"/}
    ip=${ip//\,/}
    ip=${ip//\ /}
    ip=${ip//"IPAddress:"/}

    cat << NODE >> $CONF
  - job_name: 'node_$i'
    static_configs:
      - targets: ['$ip:19090']

NODE
done

