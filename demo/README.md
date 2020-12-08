# Demo

Contains the scripts to do opera benchmarking (only for fakenet now).

## for common purpose

  - run network: `./start.sh`;
  - stop network: `./stop.sh`;
  - clean data and logs: `./clean.sh`;

You could specify number of validators by setting N environment variable.


## Stake transfer example

from [`demo/`](./demo/) dir

* Start network:
```sh
N=3 ./start.sh
```

* Attach js-console to running node0:
```sh
go run ../cmd/opera attach http://localhost:4000
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
go run ../cmd/opera attach --exec "eth.coinbase" http://localhost:4001
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
go run ../cmd/opera attach --exec "eth.getBalance(eth.coinbase)" http://localhost:4000
go run ../cmd/opera attach --exec "eth.getBalance(eth.coinbase)" http://localhost:4001
```
 outputs:
```js
9.99999999978999e+23
1.000000000000001e+24                                                                                                                                                                                       
```
