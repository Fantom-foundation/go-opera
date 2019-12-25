#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

CONF=prometheus.yml

cat << HEADER > $CONF
global:
  # How frequently to scrape targets by default.
  scrape_interval: 1m

scrape_configs:
HEADER

docker ps -f network=${NETWORK} --format '{{.Names}}' | while read svc
do
    cat << NODE >> $CONF
  - job_name: '$svc'
    static_configs:
      - targets: ['$svc:19090']
NODE
done


echo -e "\nStart Prometheus:\n"

docker run --rm -d --name=prometheus \
    --net=${NETWORK} \
    -p 9090:9090 \
    -v ${PWD}/${CONF}:/etc/prometheus/${CONF} \
    prom/prometheus
