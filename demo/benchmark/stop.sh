#!/bin/bash

PROG=lachesis

# kill all lachesis processes
pkill "${PROG}"

# remove demo data
sudo rm -rf /tmp/lachesis-demo-replay/datadir
rm -rf exec*.sh dump.traffic