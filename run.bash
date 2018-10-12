#!/usr/bin/env bash
set -e
trap 'kill $(jobs -p)' SIGINT SIGTERM EXIT

n="$1"
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PEERS_DIR="$DIR"

#PEERS_DIR="${PEERS_DIR:-$DIR}"

# [ -f "$PEERS_DIR/peers.json" ] || echo 'peers.json not found' && exit 2

digits="${#n}"

rm -f peers.json
rm -rf nodes/
batch-ethkey -dir "./nodes" -network 127.0.0.1 -n "$n" -port-start 12000 -inc-port > "$PEERS_DIR/peers.json"

node_num=0
go build -o lachesis cmd/lachesis/main.go 


for host in $(jq -rc '.[].NetAddr' "./peers.json"); do
  ip="${host%:*}";
  port="${host#*:}";
  proxy_port=$((port - 3000))
  service_port=$((port - 2000))

  printf -v node_num_p "%0${digits}d" "$node_num"
  printf '%s assigned to %s%s\n' "$host" "$PROJECT" "$node_num_p";
  echo "ip $ip"
  echo "port $port"
  echo "proxy_port $proxy_port"

  node_dir="nodes/$node_num_p"
  cp peers.json $node_dir
  ./lachesis run  --log=info --listen=$host --datadir nodes/$node_num_p --heartbeat=4s --store -p $ip:$proxy_port -s $ip:$service_port --test &


  #"$DIR/spin.bash" "$node_num_p" "$ip"
  ((node_num++))
  [ "$node_num" -gt "$n" ] && exit 0
done

wait

