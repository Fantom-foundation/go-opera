#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

n="$1"
ip_cidr="$2"
node_num=0

for ip in $(prips "$ip_cidr"); do
  ./spin.bash "$node_num" "$ip"
  ((node_num++))
  [ "$node_num" -gt "$n" ] && exit 0
done
