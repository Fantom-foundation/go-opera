#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

set -e


echo -e "\nStart $N tx-storms:\n"

METRICS=--metrics

for ((i=0;i<$N;i+=1))
do
  RPCP=$(($RPCP_BASE+$i))
  ACC=$(($i+1))
    (go run ../cmd/tx-storm \
	--num ${ACC}/$N --rate=10 \
	--accs-start=${TEST_ACCS_START} --accs-count=${TEST_ACCS_COUNT} \
	${METRICS} --verbosity 5 \
	http://127.0.0.1:${RPCP}  &> .txstorm$i.log)&
    METRICS=
done
