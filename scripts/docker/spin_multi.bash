#!/usr/bin/env bash

set -euo pipefail

IFS=$'\n\t'

declare -r DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
. "${DIR%/*}/set_globals.bash"

declare -r n="$1"
declare -i node_num=0

# [ -f "$PEERS_DIR/peers.json" ] || echo 'peers.json not found' && exit 2

declare -r digits="${#n}"

for ip in $(jq -rc '.[].NetAddr' "$PEERS_DIR/peers.json"); do
  ip="${ip%:*}";
  printf -v node_num_p "%0${digits}d" "$node_num"
  printf '%s assigned to %s%s\n' "$ip" "$PROJECT" "$node_num_p";
  "$DIR/spin.bash" "$node_num_p" "$ip"
  ((node_num++))
  [ "$node_num" -gt "$n" ] && exit 0
done
