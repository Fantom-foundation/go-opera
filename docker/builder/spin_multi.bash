#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT="${PROJECT:-lachesis}"
n="$1"
node_num=0

PEERS_DIR="${PEERS_DIR:-$BUILD_DIR}"
PEERS_DIR="${PEERS_DIR:-$DIR}"

# [ -f "$PEERS_DIR/peers.json" ] || echo 'peers.json not found' && exit 2

digits="${#n}"

for ip in $(jq -rc '.[].NetAddr' "$PEERS_DIR/peers.json"); do
  ip="${ip%:*}";
  printf -v node_num_p "%0${digits}d" "$node_num"
  printf '%s assigned to %s%s\n' "$ip" "$PROJECT" "$node_num_p";
  "$DIR/spin.bash" "$node_num_p" "$ip"
  ((node_num++))
  [ "$node_num" -gt "$n" ] && exit 0
done
