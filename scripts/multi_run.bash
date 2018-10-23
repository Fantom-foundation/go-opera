#!/usr/bin/env bash

set -xeuo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "$DIR/set_globals.bash"
. "$DIR/ncpus.bash"

# Config
declare -ri n="${n:-3}"
declare -r ip_start="${ip_start:-127.0.0.0}"
declare -r subnet="${subnet:-16}"
declare -r ip_range="$ip_start/$subnet"

# Create loopback aliases and cp json.peers per node datadir
declare -i node_num=0
for ip in $(jq -rc '.[].NetAddr' "$PEERS_DIR/lachesis_data_dir/peers.json"); do
    cp "$PEERS_DIR/lachesis_data_dir/peers.json" "$BUILD_DIR/lachesis_data_dir/$node_num/"

    ip="${ip%:*}";
    echo "$ip"
    if [ "$ip" != "127.0.0.1" ]; then
        #	sudo route add -host $ip dev lo
        loopback=$(ifconfig -a | grep lo | head -n 1 | cut -d: -f1)
    	if  ( ifconfig "$loopback":"$node_num" | grep -e 'inet 127\.' ) > /dev/null 2>& 1
	    then
            if [[ "$OSTYPE" == "linux-gnu" ]]; then
                sudo ifconfig "$loopback":"$node_num" down
            elif [[ "$OSTYPE" == "darwin"* || "$OSTYPE" == "freebsd"* ]]; then
                sudo ifconfig "$loopback" -alias "$ip"
            fi
    	fi

    	if [[ "$OSTYPE" == "linux-gnu" ]]; then
            sudo ifconfig "$loopback":"$node_num" "$ip" netmask 255.0.0.0 up
        elif [[ "$OSTYPE" == "darwin"* || "$OSTYPE" == "freebsd"* ]]; then
            sudo ifconfig "$loopback" alias "$ip" netmask 255.255.255.254 up
        fi
    else
        declare -r node_addr="$ip"
    fi
    ((node_num++)) || true
    [ "$node_num" -gt "$n" ] && exit 0
done

# Run multi lachesis
GOMAXPROCS=$(("$logicalCpuCount" - 1)) "$BUILD_DIR/lachesis_$TARGET_OS" run --datadir "$BUILD_DIR/lachesis_data_dir" --store --listen="$node_addr":12000 --log=warn --heartbeat=5s -p "$node_addr":9000 --test --test_n=10000
declare -i rc=$?
rm -rf "$BUILD_DIR/lachesis_data_dir/"
exit "$rc"
