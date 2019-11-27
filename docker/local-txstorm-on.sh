#!/usr/bin/env bash
cd $(dirname $0)

set -e

. ./_params.sh

echo -e "\nStart $N tx-storms:\n"

METRICS=--metrics

for ((i=$N-1;i>=0;i-=1))
do
  RPCP=$(($RPCP_BASE+$i))
  ACC=$(($i+1))
    (go run ../cmd/tx-storm \
	--num ${ACC}/$N \
	--rate=10 --period=30 \
	${METRICS} --verbosity 5 \
	http://127.0.0.1:${RPCP}  &> .txstorm$i.log)&
    METRICS=
done
