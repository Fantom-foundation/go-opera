#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh


echo -e "\nStart $N tx generators:\n"

for ((i=0;i<$N;i+=1))
do
    NODE=node$i
    NAME=txgen$i

    i=$((i+1))
    docker run -d --rm \
	--net=${NETWORK} --name=${NAME} \
	--cpus=${LIMIT_CPU} --blkio-weight=${LIMIT_IO} \
	tx-storm:${TAG} \
	--num=$i/$N --rate=10 --period=30 --metrics --verbosity 5 http://${NODE}:18545

done
