#!/bin/bash

PROG=opera

# kill all opera processes
pkill "${PROG}"

# remove demo data
rm -rf /tmp/opera-demo/datadir/
