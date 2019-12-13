# Docker

Contains the scripts to do lachesis benchmarking (only for fakenet now) with Docker.

## for common purpose

  - build node docker image "lachesis": `make lachesis` (use GOPROXY and TAG env vars optionally);
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

* Attach js-console to running node0:
```sh
docker exec -ti node0 /lachesis attach http://localhost:18545
```

* Check the balance to ensure that node0 has something to transfer (node0 js-console):
```js
eth.getBalance(eth.coinbase);

```
 output shows the balance value:
```js
1e+24
```

* Get node1 address:
```sh
docker exec -i node1 /lachesis attach --exec "eth.coinbase" http://localhost:18545
```
 output shows address:
```js
"0x239fa7623354ec26520de878b52f13fe84b06971"
```

* Transfer some amount from node1 to node2 address as receiver (node0 js-console):
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
docker exec -i node0 /lachesis attach --exec "eth.getBalance(eth.coinbase)" http://localhost:18545                                               
docker exec -i node1 /lachesis attach --exec "eth.getBalance(eth.coinbase)" http://localhost:18545                                               
```
 outputs:
```js
9.99999999978999e+23
1.000000000000001e+24                                                                                                                                                                                       
```


## Performance testing

[step by step example](./EXAMPLE.md)

use `cmd/tx-storm` util to generate transaction streams for each node:

  - build node docker image "tx-storm": `make tx-storm` (use GOPROXY and TAG env vars optionally);
  - start: `./txstorm-on.sh`;
  - stop: `./txstorm-off.sh`;

then collect metrics.

Also you may manually launch transactions generation only for one node - the nodes will exchange a content of their transactions pool.

## Prometheus metrics collection

  - `./prometheus-on.sh` collects metrics from running nodes and tx-storms (so run it after);
  - see webUI at `http://localhost:9090`;
  - stop: `./prometheus-off.sh`;

See results at:

 - client side: [tx latency](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_ttf&g0.tab=0)
 - client side: [count of sent txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_count_sent&g0.tab=0)
 - client side: [count of confirmed txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_count_got&g0.tab=0)
 - node side: [tx time2finish](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_tx_ttf&g0.tab=0)
 - node side: [data dir size](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_db_size&g0.tab=0)


## Testing network failures

  - install blockade and add it to "$PATH": `pip install blockade`;
  - use `./start_blockade.sh` instead normal `./start.sh`;
  - test lachesis network with blockade [`commands`](https://github.com/worstcase/blockade/blob/master/docs/commands.rst);


## without docker

You can do the same without docker, see `local-*.sh` scripts.

## FAQs

*Q: I get the following error `./_sentry.sh: line 3: declare: -l: invalid option`*

A: Try upgrading to the latest version of bash on your machine. If you are on Mac OSX, you can follow this guide [here](https://itnext.io/upgrading-bash-on-macos-7138bd1066ba).
