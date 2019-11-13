#!/bin/bash

source $(dirname $0)/set_env.sh

cat << HEADER > prometheus.yml
scrape_configs:

  - job_name: 'prometheus'
    static_configs:
      - targets: ['127.0.0.1:9090']

HEADER


for ((i=$N-1;i>=0;i-=1))                                                                                                                                                                                    
do                                                                                                                                                                                                          
    cat << SVC >> prometheus.yml
  - job_name: 'node$i'
    static_configs:
      - targets: ['$ip:19090']

  - job_name: 'txstorm$i'
    static_configs:
      - targets: ['$ip:19090']

SVC
done

docker $SWARM config create prometheus prometheus.yml
rm -f prometheus.yml


docker $SWARM service create \
  --name prometheus \
  --replicas 1 \
  --with-registry-auth \
  --detach=false \
  prom/prometheus \
    --config src=prometheus,target=/etc/prometheus/prometheus.yml \
    --publish published=9090,target=9090,mode=ingress
