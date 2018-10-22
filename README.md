# Lachesis

[documentation](http://docs.fantom.foundation) pages.


## Docker

Create an 3 node lachesis cluster with:

    n=3 BUILD_DIR="$PWD" ./scripts/docker/scale.bash

### Dependencies

  - [Docker](https://www.docker.com/get-started)
  - [jq](https://stedolan.github.io/jq)
  - [Bash](https://www.gnu.org/software/bash)
  - [git](https://git-scm.com)
  - [Go](https://golang.org)
  - [Glide](https://glide.sh)
  - [batch-ethkey](https://github.com/SamuelMarks/batch-ethkey) with: `go get -v github.com/SamuelMarks/batch-ethkey`
  - [mage](github.com/magefile/mage)
  
# LACHESIS
## BFT Consensus platform for distributed applications.

[![CircleCI](https://circleci.com/gh/mosaicnetworks/lachesis.svg?style=svg)](https://circleci.com/gh/mosaicnetworks/lachesis)

Lachesis allows many computers to behave as one. It uses Peer to Peer (P2P) 
networking and a consensus algorithm to guarantee that multiple connected 
computers process the same commands in the same order; a technique known as 
state machine replication. This makes for secure systems that can tolerate 
arbitrary failures including malicious behavior.

For guidance on how to install and use Lachesis please visit our 
[documentation](http://docs.lachesis.io) pages.

**NOTE**:
This is alpha software. Please contact us if you intend to run it in production.

## Consensus Algorithm and Blockchain

We use an adaptation of the Poset consensus algorithm, invented by Leemon 
Baird. Poset is best described in the 
[white-paper](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf) and its 
[accompanying document](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-02.pdf).
The original algorithm is protected by [patents](http://www.swirlds.com/ip/) in 
the USA. Therefore, anyone intending to use this software in the USA should 
obtain a license from the patent holders.

Poset is based on the intuitive idea that gossiping about gossip itself 
yields enough information to compute a consensus ordering of events. It attains 
the theoretical limit of tolerating up to one-third of faulty nodes without 
compromising on speed. For those familiar with the jargon, it is a leaderless, 
asynchronous BFT consensus algorithm.

Lachesis projects the output of the consensus algorithm onto a linear blockchain 
which is more suitable for representing an ordered list of transactions and 
facilitates the creation of light-clients. For information about this projection
please refer to [documentation](http://docs.lachesis.io/blockchain.html) pages.

## Design

Lachesis is designed to integrate with applications written in any programming 
language.

### Overview

```
    +--------------------------------------+
    | APP                                  |
    |                                      |
    |  +-------------+     +------------+  | 
    |  | Service     | <-- | State      |  |
    |  |             |     |            |  |
    |  +-------------+     +------------+  |
    |          |                ^          |
    |          |                |          |
    +----------|----------------|----------+
               |                |                      
--------- SubmitTx(tx) ---- CommitBlock(Block) ------- JSON-RPC/TCP or in-memory       
               |                |                         
 +-------------|----------------|------------------------------+
 | LACHESIS      |                |                              |
 |             v                |                              |
 |          +----------------------+                           |
 |          | App Proxy            |                           |
 |          |                      |                           |
 |          +----------------------+                           |
 |                     |                                       |
 |   +-------------------------------------+                   |
 |   | Core                                |                   |
 |   |                                     |                   |
 |   |  +------------+                     |    +----------+   |
 |   |  | Poset  |       +---------+   |    | Service  |   |
 |   |  +------------+       | Store   |   | -- |          | <----> HTTP 
 |   |  +------------+       +----------   |    |          |   |
 |   |  | Blockchain |                     |    +----------+   |
 |   |  +------------+                     |                   |
 |   |                                     |                   |
 |   +-------------------------------------+                   |
 |                     |                                       |
 |   +-------------------------------------+                   |
 |   | Transport                           |                   |
 |   |                                     |                   |
 |   +-------------------------------------+                   |
 |                     ^                                       |
 +---------------------|---------------------------------------+
                       |
                       v
                  P2P Network
```

Almost any software application can be modeled in terms of a *service* and a 
*state*. The *service* is responsible for processing commands (ex. user input), 
while the *state* is responsible for manipulating and storing the data (eg. 
database). Usually, when commands need to update the data, the *service* will 
invoke the *state* directly. In a distributed application, however, commands 
(referred to as *transactions* in this context), must be broadcast to all 
replicas and consistently ordered before being applied to the *state*. This 
ensures that all replicas process the same commands in the same order. Hence, 
the *service* no longer communicates directly with the *state* (except for 
read-only requests), but forwards commands to a *transaction ordering system* 
which takes care of broadcasting and ordering the transactions across all 
replicas before feeding them back to the application's *state*. 

Lachesis is an ordering system that plugs into any application thanks to a very 
simple interface. It uses a consensus algorithm, to replicate and order the 
transactions, and a blockchain to represent the resulting list. A blockchain is 
a linear data structure composed of batches of transactions, hashed and signed 
together, easily allowing to verify any transaction. So, instead of applying 
commands directly to the *state*, Lachesis applications must forward the commands 
to Lachesis and let them be processed asynchronously by the consensus system 
before receiving them back, in blocks, ready to be applied to the *state*.  

### API

Lachesis communicates with the App through an `AppProxy` interface, which has two
implementations:

- `SocketProxy`: A SocketProxy connects to an App via TCP sockets. It enables 
                 the application to run in a separate process or machine, and to 
                 be written in any programming language.

- `InmemProxy` : An InmemProxy uses native callback handlers to integrate Lachesis 
                 as a regular Go dependency. 

The `AppProxy` interface exposes three methods for Lachesis to call the App:

- `CommitBlock(Block) ([]byte, error)`: Commits a block to the application and 
                                        returns the resulting state hash.
- `GetSnapshot(int) ([]byte, error)`: Gets the application snapshot 
                                      corresponding to a particular block index.
- `Restore([]byte) error`: Restores the App state from a snapshot.

Reciprocally, `AppProxy` relays transactions from the App to Lachesis via a native 
Go channel - `SubmitCh` - which ties into the application differently depending 
on the type of proxy (Socket or Inmem).

Lachesis asynchronously processes transactions and eventually feeds them back to 
the App, in consensus order and bundled into blocks, with a **CommitBlock** 
call. Transactions are just raw bytes and Lachesis does not need to know what they 
represent. Therefore, encoding and decoding transactions is done by the App.

Refer to the [API documentation](http://docs.lachesis.io/latest/api.html) for more 
details and a quick-start guide.

## Dev

### Go
Lachesis is written in [Golang](https://golang.org/). Hence, the first step is to 
install **Go version 1.9 or above** which is both the programming language and a 
CLI tool for managing Go code. Go is very opinionated and will require you to 
[define a workspace](https://golang.org/doc/code.html#Workspaces) where all your 
go code will reside.

### Lachesis and dependencies
Clone the [repository](https://github.com/mosaicnetworks/lachesis) in the appropriate 
GOPATH subdirectory:

```bash
$ mkdir -p $GOPATH/src/github.com/mosaicnetworks/
$ cd $GOPATH/src/github.com/mosaicnetworks
[...]/mosaicnetworks$ git clone https://github.com/mosaicnetworks/lachesis.git
```
Lachesis uses [Glide](http://github.com/Masterminds/glide) to manage dependencies.

```bash
[...]/lachesis$ curl https://glide.sh/get | sh
[...]/lachesis$ glide install
```
This will download all dependencies and put them in the **vendor** folder.

### Other requirements

Bash scripts used in this project assume the use of GNU versions of coreutils.
Please ensure you have GNU versions of these programs installed:-

example for macos:
```
# --with-default-names makes the `sed` and `awk` commands default to gnu sed and gnu awk respectively.
brew install gnu-sed gawk --with-default-names
```

### Testing

Lachesis has extensive unit-testing. Use the Go tool to run tests:
```bash
[...]/lachesis$ make test
```

If everything goes well, it should output something along these lines:
```
?       github.com/mosaicnetworks/lachesis/src/lachesis     [no test files]
ok      github.com/mosaicnetworks/lachesis/src/common     0.015s
ok      github.com/mosaicnetworks/lachesis/src/crypto     0.122s
ok      github.com/mosaicnetworks/lachesis/src/poset  10.270s
?       github.com/mosaicnetworks/lachesis/src/mobile     [no test files]
ok      github.com/mosaicnetworks/lachesis/src/net        0.012s
ok      github.com/mosaicnetworks/lachesis/src/node       19.171s
ok      github.com/mosaicnetworks/lachesis/src/peers      0.038s
?       github.com/mosaicnetworks/lachesis/src/proxy      [no test files]
ok      github.com/mosaicnetworks/lachesis/src/proxy/dummy        0.013s
ok      github.com/mosaicnetworks/lachesis/src/proxy/inmem        0.037s
ok      github.com/mosaicnetworks/lachesis/src/proxy/socket       0.009s
?       github.com/mosaicnetworks/lachesis/src/proxy/socket/app   [no test files]
?       github.com/mosaicnetworks/lachesis/src/proxy/socket/lachesis        [no test files]
?       github.com/mosaicnetworks/lachesis/src/service    [no test files]
?       github.com/mosaicnetworks/lachesis/src/version    [no test files]
?       github.com/mosaicnetworks/lachesis/cmd/lachesis     [no test files]
?       github.com/mosaicnetworks/lachesis/cmd/lachesis/commands    [no test files]
?       github.com/mosaicnetworks/lachesis/cmd/dummy      [no test files]
?       github.com/mosaicnetworks/lachesis/cmd/dummy/commands     [no test files]

```

## Build from source

The easiest way to build binaries is to do so in a hermetic Docker container. 
Use this simple command:

```bash
[...]/lachesis$ make dist
```
This will launch the build in a Docker container and write all the artifacts in
the build/ folder.

```bash
[...]/lachesis$ tree build
build/
├── dist
│   ├── lachesis_0.1.0_darwin_386.zip
│   ├── lachesis_0.1.0_darwin_amd64.zip
│   ├── lachesis_0.1.0_freebsd_386.zip
│   ├── lachesis_0.1.0_freebsd_arm.zip
│   ├── lachesis_0.1.0_linux_386.zip
│   ├── lachesis_0.1.0_linux_amd64.zip
│   ├── lachesis_0.1.0_linux_arm.zip
│   ├── lachesis_0.1.0_SHA256SUMS
│   ├── lachesis_0.1.0_windows_386.zip
│   └── lachesis_0.1.0_windows_amd64.zip
└── pkg
    ├── darwin_386
    │   └── lachesis
    ├── darwin_amd64
    │   └── lachesis
    ├── freebsd_386
    │   └── lachesis
    ├── freebsd_arm
    │   └── lachesis
    ├── linux_386
    │   └── lachesis
    ├── linux_amd64
    │   └── lachesis
    ├── linux_arm
    │   └── lachesis
    ├── windows_386
    │   └── lachesis.exe
    └── windows_amd64
        └── lachesis.exe
```

## Demo

To see Lachesis in action, we have provided a series of scripts to bootstrap a 
test network locally.

**NOTE:**
This has been tested on Ubuntu 16.04 and macOS.

Make sure you have [Docker](https://docker.com) installed.

Then, run the testnet:

```bash
[...]/lachesis$ cd demo
[...]/lachesis/demo$ make
```

Once the testnet is started, a script is automatically launched to monitor 
consensus figures:

```
consensus_events:180 consensus_transactions:40 events_per_second:0.00 id:1 last_block_index:3 last_consensus_round:17 num_peers:3 round_events:7 rounds_per_second:0.00 state:Babbling sync_rate:1.00 transaction_pool:0 undetermined_events:18
consensus_events:180 consensus_transactions:40 events_per_second:0.00 id:3 last_block_index:3 last_consensus_round:17 num_peers:3 round_events:7 rounds_per_second:0.00 state:Babbling sync_rate:1.00 transaction_pool:0 undetermined_events:20
consensus_events:180 consensus_transactions:40 events_per_second:0.00 id:2 last_block_index:3 last_consensus_round:17 num_peers:3 round_events:7 rounds_per_second:0.00 state:Babbling sync_rate:1.00 transaction_pool:0 undetermined_events:21
consensus_events:180 consensus_transactions:40 events_per_second:0.00 id:0 last_block_index:3 last_consensus_round:17 num_peers:3 round_events:7 rounds_per_second:0.00 state:Babbling sync_rate:1.00 transaction_pool:0 undetermined_events:20
```

Running ```docker ps -a``` will show you that 9 docker containers have been launched:
```
[...]/lachesis/demo$ docker ps -a
CONTAINER ID        IMAGE                    COMMAND                  CREATED             STATUS              PORTS                   NAMES
ba80ef275f22        mosaicnetworks/watcher   "/watch.sh"              48 seconds ago      Up 7 seconds                                watcher
4620ed62a67d        mosaicnetworks/dummy     "dummy '--name=client"   49 seconds ago      Up 48 seconds       1339/tcp                client4
847ea77bd7fc        mosaicnetworks/lachesis    "lachesis run --cache_s"   50 seconds ago      Up 49 seconds       80/tcp, 1337-1338/tcp   node4
11df03bf9690        mosaicnetworks/dummy     "dummy '--name=client"   51 seconds ago      Up 50 seconds       1339/tcp                client3
00af002747ca        mosaicnetworks/lachesis    "lachesis run --cache_s"   52 seconds ago      Up 50 seconds       80/tcp, 1337-1338/tcp   node3
b2011d3d65bb        mosaicnetworks/dummy     "dummy '--name=client"   53 seconds ago      Up 51 seconds       1339/tcp                client2
e953b50bc1db        mosaicnetworks/lachesis    "lachesis run --cache_s"   53 seconds ago      Up 52 seconds       80/tcp, 1337-1338/tcp   node2
0c9dd65de193        mosaicnetworks/dummy     "dummy '--name=client"   54 seconds ago      Up 53 seconds       1339/tcp                client1
d1f4e5008d4d        mosaicnetworks/lachesis    "lachesis run --cache_s"   55 seconds ago      Up 54 seconds       80/tcp, 1337-1338/tcp   node1
```
Indeed, each node is comprised of an App and a Lachesis node (cf Design section).

Run the **demo** script to play with the **Dummy App** which is a simple chat 
application powered by the Lachesis consensus platform:

```
[...]/lachesis/demo$ make demo
```
![Demo](demo/img/demo.png)


Finally, stop the testnet:
```
[...]/lachesis/demo$ make stop
```