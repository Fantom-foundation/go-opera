#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

docker $SWARM config create prometheus - < <(

cat << HEADER
scrape_configs:

HEADER

for MASK in node txgen
do
    for NAME in $(docker $SWARM service ls --filter "name=${MASK}" --format "{{.Name}}")
    do
        cat << SVC
  - job_name: '${NAME}'
    static_configs:
      - targets: ['${NAME}:19090']

SVC
    done
done

)

#docker $SWARM config create prometheus prometheus.yml
#rm -f prometheus.yml

docker $SWARM service create \
  --network lachesis \
  --hostname="{{.Service.Name}}" \
  --name prometheus \
  --config src=prometheus,target=/etc/prometheus/prometheus.yml \
  --publish 9090:9090 \
  --replicas 1 \
  --with-registry-auth \
  --detach=false \
  prom/prometheus

#  --publish published=9090,target=9090,mode=ingress \
