#!/usr/bin/env bash

#declare -ri N="${N:-3}"
declare -ri N="${N:-4}"
declare -ri M="${M:-2}"
declare -r  TAG="${TAG:-latest}"

PORT_BASE=3000
RPCP_BASE=4000
WSP_BASE=4500

attach_and_exec() {
    local i=$1
    local CMD=$2
    local RPCP=$(($RPCP_BASE+$i))

    for attempt in $(seq 40)
    do
        if (( attempt > 5 ))
        then
            echo "  - attempt ${attempt}: " >&2
        fi

        res=$(../build/demo_opera --exec "${CMD}" attach http://127.0.0.1:${RPCP} 2> /dev/null)
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
