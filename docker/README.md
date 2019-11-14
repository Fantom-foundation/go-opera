# Docker

Contains the scripts to do lachesis benchmarking (only for fakenet now) with Docker.

## for common purpose

  - build node docker image "lachesis": `make lachesis` (use GOPROXY env optionally);
  - run network: `./start.sh`;
  - stop network: `./stop.sh`;

You could specify number of validators by setting N environment variable.
It is possible to get the error "failed RPC connection to" because nodes start slowly. Try `./start.sh` again.


## the same with Sentry service

  - set `SENTRY=yes` before `./start.sh`;
  - remove Sentry database: `sentry/clean.sh`;

After startup go to http://localhost:9000 and sign in using your Sentry-account to watch and manage logs from all running local nodes.
Logs are grouped and colored (info - blue, warn - yellow, error - red).
Each log includes: environment info, message about an error, code line (in case of an error).


## Stake transfer example

from [`docker/`](./docker/) dir

* Start network:
```sh
N=3 ./start.sh
```

* Attach js-console to running node1:
```sh
docker exec -ti lachesis-node-1 /lachesis attach http://localhost:18545
```

* Check the balance to ensure that node1 has something to transfer (node1 js-console):
```js
eth.getBalance(eth.coinbase);

```
 output shows the balance value:
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
 output shows unique hash of the outgoing transaction:
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

* As soon as transaction is included into a block you will see new balance of both node addresses:
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

use `cmd/tx-storm` util to generate transaction streams for each node:

  - build node docker image "tx-storm": `make tx-storm`;
  - start: `./txstorm-on.sh`;
  - stop: `./txstorm-off.sh`;

then collect metrics.

Also you may manually launch transactions generation only for one node - the nodes will exchange a content of their transactions pool.

## Prometheus metrics collection

  - `./prometheus-on.sh` collects metrics from running nodes and tx-storms (so run it after);
  - see webUI at `http://localhost:9090`;
  - stop: `./prometheus-off.sh`;

See results at:

 - [tx latency](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_tx_latency&g0.tab=0)
 - [count of sent txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_tx_count_sent&g0.tab=0)
 - [count of confirmed txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_tx_count_sent&g0.tab=0)


## Testing network failures

  - install blockade and add it to "$PATH": `pip install blockade`;
  - use `./start_blockade.sh` instead normal `./start.sh`;
  - test lachesis network with blockade [`commands`](https://github.com/worstcase/blockade/blob/master/docs/commands.rst);


## without docker

You can do the same without docker, see `local-*.sh` scripts.
