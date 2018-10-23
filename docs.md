# Lachesis Code Documentation

## Code organization

The code itself is a fork of [Babble](https://github.com/andrecronje/babble) ([docs](http://babbleio.readthedocs.io/en/latest/)). Babble is a pretty straightforward go implementation of [Hashgraph](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf). The hashgraph paper is well written, and recommended reading.

Hashgraph has the following core concepts:

  - **Transactions**: Transactions are binary data chunks of application logic. The binary chunks encapsulate state changes - which eventually will include payments and things like that. Nothing in the codebase interprets transactions yet.
  - **Events**: Events are the core unit of concurrency for Hashgraph. Events are made of a group of transactions. Each event is generated on a single server, signed, and passed between servers using the hashgraph gossip protocol. Events are defined in `src/poset/event.go`.
  - **Nodes**: A node is a server in the hashgraph consensus algorithm. There is no mechanism to add or remove nodes. All nodes must be known to all participants when the consensus group is initially created.

The code is roughly organised as follows:

  - `/cmd/lachesis/main.go`: Main entrypoint
  - `/src/*`: Library code
    - `/src/node`: The core runtime / state machine implementation. `node/node.go` and `node/core.go` are the heart of the machine.
    - `/src/poset`: Actual hashgraph algorithm implementation. Most of the interesting logic happens here. This is [forked from `/hashgraph` in babble](https://github.com/andrecronje/babble/tree/master/hashgraph).
    - `/src/proxy`: Proxy is a misnomer. This provides 2 APIs:
      - `/src/proxy/lachesis/*` is an API for 3rd party clients to talk to the server and commit transactions into the network
      - `/src/proxy/app/*` is an API for application logic & state transitions to exist out of process (for example account balances and stuff).
    - `/src/service`: This is a REST HTTP service for writing block explorers and stuff like that
  - `/tester/tester.go`: Some code to inject transactions into the server for benchmarking
  - `/docker/builder/*.bash`: Scripts to run the thing through docker.

## Data flow

The flow of a transaction through the system is as follows:

  1. A transaction is created by a client and submitted over a JSON RPC server listening on the proxy port (defaults to localhost:1338 but configurable with `-p ip:port`). The `SocketLachesisProxyClient` interface exposes this functionality.
  2. The transaction is forwarded via the `SocketAppProxyServer#SubmitTx` to the core node process in `src/node/node.go`.
  3. The transaction is sent to the poset code - which manages the core hashgraph algorithm. Here it is added to the transaction mempool.
  4. Eventually all transactions in the mempool will be merged into an Event object. This happens every heartbeat - which is configured by passing `--heartbeat=4s` to the process.
  5. The event object is passed through the [hashgraph consensus algorithm](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf). Read [the paper](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf) - its really good.
  6. Once an event has been seen by 2n/3+1 nodes, it is considered to be committed. In the code it is reffered to as a 'consensus event' containing 'consensus transactions'. Only at this point in the process does the event mutate the 'application state' (balances, smart contract state, etc). So, if a 
transaction moves money from wallet 1 to wallet 2, it is only at this point in the algorithm that the money will actually be visibly transferred between wallets.
  7. Any lachesis proxy clients are told about the transaction here.

Most of the heart of the whole system is the innocuously named `Node#doBackgroundWork()` function in `src/node/node.go`:

```go
func (n *Node) doBackgroundWork() {
	for {
		select {
		case rpc := <-n.netCh:
			// Process server <-> server comms
		case block := <-n.commitCh:
			// Process new event block from hashgraph
		case t := <-n.submitCh:
			// Forward transaction from client -> hashgraph
		case <-n.shutdownCh:
			return
		}
	}
}
```

## Store

The events and transactions are stored either in an in memory database or an on disk KV database ([badger](https://github.com/dgraph-io/badger)). If you use badger, the in-memory store is still used as an LRU cache for recent events.

Enable badger by passing `--store` at startup.

## Running the lachesis server

#### Running locally

The simplest way to run the server is with `go run`. By default this will look for a peers.json file and a keypair in `~/.lachesis`.

```bash
$ go run cmd/lachesis/main.go run  --log=debug --listen=127.0.0.1:12000 --heartbeat=10s --store
```

You can generate all these files with a script...

#### Running a local cluster

#### Running through docker

You can build & run a cluster of docker instances like this:

```bash
n=3 BUILD_DIR="$PWD" ./scripts/docker/scale.bash
```

#### Command line arguments

> **TODO**
