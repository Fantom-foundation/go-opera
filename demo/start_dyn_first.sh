#!/bin/bash

# This script will launch a cluster of Lachesis nodes
# The parameter N = number of nodes to run

. ./utils.sh

# create and start nodes
start_nodes 1 $N $T

# connect these nodes together
connect_nodes 1 $N

echo
echo "Sleep for 10000 seconds..."
sleep 10000
