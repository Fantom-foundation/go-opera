#!/usr/bin/env bash

CONF=prometheus.yml

cat << HEADER > $CONF
scrape_configs:

  - job_name: 'prometheus'
    static_configs:
      - targets: ['127.0.0.1:9090']
HEADER


docker ps -f network=${NETWORK} --format '{{.Names}}' | while read svc
do
    ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${svc})

    cat << NODE >> $CONF
  - job_name: '$svc'
    static_configs:
      - targets: ['$ip:19090']
NODE
done


echo -e "\nStart Prometheus:\n"

docker run --rm -d --name=prometheus \
    --net=${NETWORK} \
    -p 9090:9090 \
    -v ${PWD}/${CONF}:/etc/prometheus/${CONF} \
    prom/prometheus
