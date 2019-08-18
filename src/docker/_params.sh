#!/usr/bin/env bash

declare -ri N="${N:-3}"

NAME=pos-lachesis-node

LIMIT_CPU=$(echo "scale=2; 1/$N" | bc)
LIMIT_IO=$(echo "500/$N" | bc)
