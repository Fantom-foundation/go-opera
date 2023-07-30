# Demo

This directory contains the scripts to run fakenet (private testing network) with N+M local nodes,
primarily for benchmarking purposes.

## Scripts

  - start network: `./start.sh`;
  - stop network: `./stop.sh`;
  - run new validator node: `./run-validator.sh 3`;
  - clean data and logs: `./clean.sh`;

You can specify number of genesis validators by setting N environment variable
 and M for additional nodes.

## Balance transfer example

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
ftm.getBalance(ftm.accounts[0]);
```
 
 output shows the balance value:
```js
1e+27
```

* Get node1 address:
```sh
go run ../cmd/opera attach --exec "ftm.accounts[0]" http://localhost:4001
```
 output shows address:
```js
"0x02aff1d0a9ed566e644f06fcfe7efe00a3261d03"
```

* Transfer some amount from node0 to node1 address as receiver (node0 js-console):
```js
ftm.sendTransaction(
	{from: ftm.accounts[0], to: "0x02aff1d0a9ed566e644f06fcfe7efe00a3261d03", value:  "1000000000"},
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
ftm.getTransactionReceipt("0x68a7c1daeee7e7ab5aedf0d0dba337dbf79ce0988387cf6d63ea73b98193adfd").blockNumber
```
 output shows number of block, transaction was included in:
```
174
```

* As soon as transaction is included into a block you will see new balance of both node addresses:
```sh
go run ../cmd/opera attach --exec "ftm.getBalance(ftm.accounts[0])" http://localhost:4000
go run ../cmd/opera attach --exec "ftm.getBalance(ftm.accounts[0])" http://localhost:4001
```
 outputs:
```js
9.99999999978999e+26
1.000000000000001e+27
```
