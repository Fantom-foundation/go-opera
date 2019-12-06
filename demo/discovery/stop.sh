#!/bin/bash

# kill all bootnode and lachesis processes
pkill "bootnode"
pkill "lachesis"

# remove demo data
#rm -rf /tmp/lachesis-demo/datadir/
