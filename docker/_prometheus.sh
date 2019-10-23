#!/usr/bin/env bash
CONF=prometheus.yml

declare -rl PROMETHEUS="${PROMETHEUS:-no}"
if [ "$PROMETHEUS" == "yes" ]
then
    echo -e "\nStart Prometheus:\n"

    cat << HEADER > $CONF
scrape_configs:

  - job_name: 'prometheus'
    static_configs:
      - targets: ['127.0.0.1:9090']
HEADER


    docker ps -f name=${NAME} --format '{{.Names}}' | while read n
    do
	ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $n)

        cat << NODE >> $CONF
  - job_name: '$n'
    static_configs:
      - targets: ['$ip:19090']
NODE
    done

    docker run --rm -d --name=prometheus \
	--net=${NETWORK} \
	-p 9090:9090 \
	-v ${PWD}/${CONF}:/etc/prometheus/${CONF} \
	prom/prometheus
                                                                                                                                            
fi
