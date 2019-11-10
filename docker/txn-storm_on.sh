#!/usr/bin/env bash
cd $(dirname $0)

. ./_params.sh


n=$(docker ps -a -q -f network=${NETWORK} -f name=${NAME} | wc -l)
echo -e "\nStart $N txn generators:\n"
i=0

docker ps -f network=${NETWORK} -f name=${NAME} --format '{{.Names}}' | while read node
do
    ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${node})
    i=$((i+1))
    docker run -d --rm \
	--net=${NETWORK} --name=txngen-$i \
	--cpus=${LIMIT_CPU} --blkio-weight=${LIMIT_IO} \
	"txn-storm" \
	--num $i/$n --donor=$i --rate=10 --period=60 http://${ip}:18545

done
