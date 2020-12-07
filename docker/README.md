# Docker

Contains the scripts to do opera benchmarking (only for fakenet now) with Docker.

## for common purpose

  - build node docker image "opera": `make opera` (use GOPROXY and TAG env vars optionally);
  - run network: `./start.sh`;
  - stop network: `./stop.sh`;

You could specify number of validators by setting N environment variable.
It is possible to get the error "failed RPC connection to" because nodes start slowly. Try `./start.sh` again.


## Stake transfer example

from [`docker/`](./docker/) dir

* Start network:
```sh
N=3 ./start.sh
```

* Attach js-console to running node0:
```sh
docker exec -ti node0 /opera attach http://localhost:18545
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
docker exec -i node1 /opera attach --exec "eth.coinbase" http://localhost:18545
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
docker exec -i node0 /opera attach --exec "eth.getBalance(eth.coinbase)" http://localhost:18545                                               
docker exec -i node1 /opera attach --exec "eth.getBalance(eth.coinbase)" http://localhost:18545                                               
```
 outputs:
```js
9.99999999978999e+23
1.000000000000001e+24                                                                                                                                                                                       
```


## without docker

You can do the same without docker, see `local-*.sh` scripts.

