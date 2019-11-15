#!/usr/bin/env bash
cd $(dirname $0)

. ./_params.sh


n=$(docker ps -a -q -f network=${NETWORK} -f name=${NAME} | wc -l)
echo -e "\nStart $N tx generators:\n"
i=0

docker ps -f network=${NETWORK} -f name=${NAME} --format '{{.Names}}' | while read node
do
    ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${node})
    i=$((i+1))
    docker run -d --rm \
	--net=${NETWORK} --name=txgen-$i \
	--cpus=${LIMIT_CPU} --blkio-weight=${LIMIT_IO} \
	"tx-storm" \
	--num=$i/$n --rate=10000 --accs=${TEST_ACCS} --metrics --verbosity 5 http://${ip}:18545

done
