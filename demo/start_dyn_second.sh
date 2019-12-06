#!/bin/bash

# This script will launch a cluster of Lachesis nodes

. ./utils.sh

# start new nodes
echo "Starting the second group of nodes"
echo -e "\nStart $M nodes:\n"
for i in $(seq $((N+1)) $T);
do
    start_node $i $T
done

# and connect the new nodes together with the existing ones
echo -e "\nConnect new nodes to ring:\n"
for i in $(seq $((N+1)) $((T-1)));
do
    j=$((i + 1))
    echo "i=$i, j=$j"
    conn=$(connect_pair $i $j)
done

conn=$(connect_pair 1 $((N+1)))
conn=$(connect_pair $T 1)

##


echo
echo "Sleep for 10000 seconds..."
sleep 10000
