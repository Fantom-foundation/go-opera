#!/usr/bin/env bash -euo pipefail

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

. "$DIR/set_globals.bash"

declare -ri n="$1"

if [ -z "$n" ]; then
   >&2 echo "Usage: $0 <number_of_nodes>"
   exit 1
fi

# [ -f "$PEERS_DIR/peers.json" ] || echo 'peers.json not found' && exit 2

declare -ri digits="${#n}"

rm -rf "$BUILD_DIR/nodes" "$BUILD_DIR/peers.json"
batch-ethkey -dir "$PEERS_DIR/nodes" -network 127.0.0.1 -n "$n" -port-start 12000 -inc-port > "$PEERS_DIR/peers.json"


declare -i node_num=0
go build -o lachesis cmd/lachesis/main.go

trap 'kill $(jobs -p)' SIGINT SIGTERM EXIT

for host in $(jq -rc '.[].NetAddr' "$PEERS_DIR/peers.json"); do
  ip="${host%:*}";
  port="${host#*:}";
  declare -r proxy_port=$((port - 3000))
  declare -r service_port=$((port - 2000))

  printf -v node_num_p "%0${digits}d" "$node_num"
  printf '%s assigned to %s%s\n' "$host" "$PROJECT" "$node_num_p";
  echo "ip $ip"
  echo "port $port"
  echo "proxy_port $proxy_port"

  node_dir="$PEERS_DIR/nodes/$node_num_p"
  cp "$PEERS_DIR/peers.json" "$node_dir"
  ./lachesis run  --log=info --listen="$host" --datadir "$PEERS_DIR/nodes/$node_num_p" --heartbeat=4s --store -p "$ip:$proxy_port" -s "$ip:$service_port" --test &

  #"$DIR/spin.bash" "$node_num_p" "$ip"
  ((node_num++))
  [ "$node_num" -gt "$n" ] && exit 0
done

wait
