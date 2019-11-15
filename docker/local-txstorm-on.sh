#!/usr/bin/env bash
cd $(dirname $0)

set -e

. ./_params.sh

echo -e "\nStart $N tx-storms:\n"

METRICS=--metrics

for i in $(seq $N)
do
    (go run ../cmd/tx-storm \
	--num $i/$N \
	--rate=30000 --accs=${TEST_ACCS} \
	${METRICS} --verbosity 5 \
	http://127.0.0.1:$((4000+i))  &> .txstorm$i.log)&
    METRICS=
done
