#!/bin/bash

PROG=lachesis

# kill all lachesis processes
pkill "${PROG}"

# remove demo data
rm -rf /tmp/lachesis-demo/datadir/
