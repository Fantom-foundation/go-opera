# Docker

Contains scripts to try lachesis (only for fakenet now) with Docker.


## for common purpose

  - build node docker image "lachesis": `make lachesis`;
  - run network of x nodes: `N=x ./start.sh`;
  - stop network: `./stop.sh`;


## the same with Sentry service

  - set `SENTRY=yes` before `./start.sh`;
  - remove Sentry database: `sentry/clean.sh`;

During first run, you'll get offer to create Sentry-account. Note that account exist only inside local copy Sentry and don't affect to sentry.io.
After start up go to `http://localhost:9000` and sign in using that Sentry-account to see and management logs from all running local nodes.
Logs are grouped and marked with color (info - blue, warn - yellow, error - red).
Each log include: environment info, message about error, code line (in case error).


## Prometheus metrics collection:

  - collect metrics from runnings: `./prometheus_on.sh` (after `./start.sh`);
  - see webUI at `http://localhost:9090`;
  - stop: `./prometheus_off.sh`;


## Stake transfer example

from [`docker/`](./docker/) dir

* Start network:
```sh
N=3 ./start.sh
```

* Attach js-console of running node1:
```sh
docker exec -ti lachesis-node-1 /lachesis attach http://localhost:18545
```

* Check the stake to ensure the node1 have something to transfer (node1 js-console):
```js
eth.getBalance(eth.coinbase);

```
 output shows stake value:
```js
1e+24
```

* Get node2 address:
```sh
docker exec -i lachesis-node-2 /lachesis attach --exec "eth.coinbase" http://localhost:18545
```
 output shows address:
```js
"0x239fa7623354ec26520de878b52f13fe84b06971"
```

* Transfer some amount from node1 to node2 address as receiver (node1 js-console):
```js
eth.sendTransaction(
	{from: eth.coinbase, to: "0x239fa7623354ec26520de878b52f13fe84b06971", value:  "1000000000"},
	function(err, transactionHash) {
        if (!err)
            console.log(transactionHash + " success"); 
    });
```
 output shows unique hash the outgoing transaction:
```js
0x68a7c1daeee7e7ab5aedf0d0dba337dbf79ce0988387cf6d63ea73b98193adfd success
```

* Check the transaction status by its unique hash (js-console):
```sh
eth.getTransactionReceipt("0x68a7c1daeee7e7ab5aedf0d0dba337dbf79ce0988387cf6d63ea73b98193adfd").blockNumber
```
 output shows number of block, transaction was included in:
```
174
```

* As soon as transaction included into block you will see new stake of both nodes:
```sh
docker exec -i lachesis-node-1 /lachesis attach --exec "eth.getBalance(eth.coinbase)" http://localhost:18545                                               
docker exec -i lachesis-node-2 /lachesis attach --exec "eth.getBalance(eth.coinbase)" http://localhost:18545                                               
```
 outputs:
```js
9.99999999978999e+23
1.000000000000001e+24                                                                                                                                                                                       
```


## Performance testing

use `cmd/tx-storm` util to generate transaction streams:

  - start: `./tx-storm_on.sh`;
  - stop: `./tx-storm_off.sh`;

and Prometheus to collect metrics.


## Testing network failures

  - install blockade and add it to "$PATH": `pip install blockade`;
  - use `./start_blockade.sh` instead normal `./start.sh`;
  - test lachesis network with blockade [`commands`](https://github.com/worstcase/blockade/blob/master/docs/commands.rst);
