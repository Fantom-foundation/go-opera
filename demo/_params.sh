#!/usr/bin/env bash

declare -ri N="${N:-3}"
declare -ri M="${M:-0}"
declare -r  TAG="${TAG:-latest}"

PORT_BASE=3000
RPCP_BASE=4000
WSP_BASE=4500

data_dir() {
	local i=$1
	echo "${PWD}/opera$i.datadir"
}

run_opera_node() {
	local i=$1
	local ACC=$(($i+1))
	local DATADIR="$(data_dir $i)"
	local PORT=$(($PORT_BASE+$i))
	local RPCP=$(($RPCP_BASE+$i))
	local WSP=$(($WSP_BASE+$i))

	../build/demo_opera \
		--datadir=${DATADIR} \
		--fakenet=${ACC}/$N \
		--port=${PORT} \
		--nat extip:127.0.0.1 \
		--http --http.addr="127.0.0.1" --http.port=${RPCP} --http.corsdomain="*" --http.api="eth,debug,net,admin,web3,personal,txpool,ftm,dag" \
		--ws --ws.addr="127.0.0.1" --ws.port=${WSP} --ws.origins="*" --ws.api="eth,debug,net,admin,web3,personal,txpool,ftm,dag" \
		--metrics --metrics.addr=127.0.0.1 --metrics.port=$(($RPCP+1100)) \
		--verbosity=3 --tracing >> opera$i.log 2>&1
}

attach_and_exec() {
    local i=$1
    local DATADIR="$(data_dir $i)"
    local CMD=$2    

    for attempt in $(seq 40)
    do
        if (( attempt > 5 ))
        then
            echo "  - attempt ${attempt}: " >&2
        fi

        res=$(../build/demo_opera --exec "${CMD}" attach ${DATADIR}/opera.ipc 2> /dev/null)
        if [ $? -eq 0 ]
        then
            #echo "success" >&2
            echo $res
            return 0
        else
            #echo "wait" >&2
            sleep 1
        fi
    done
    echo "failed RPC connection to ${NAME}" >&2
    return 1
}
