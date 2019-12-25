#!/bin/bash

echo txgen0
./swarm service logs --tail 50000 txgen0 2> scope3/txstorm0.log

for ((i=$N-1;i>=0;i-=1))
do
    echo node$i
    ./swarm service logs --tail 50000 node$i 2> scope3/node$i.log
done

#ftm.sendTransaction({from: ftm.coinbase, to: "0x2511a9150f2aa355900b8ae13fdc59d1c91bcdd5", value:1e21});
#ftm.sendTransaction({from: ftm.coinbase, to: "0x2e6310b13984ad44da15aa5e3a78e57ed2bf490c", value:1e21});
#ftm.sendTransaction({from: ftm.coinbase, to: "0x074e79af2998F4C62f9e19Ea8fB84cE209decF85", value:4e24});
