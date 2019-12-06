# Introduction

The `demo` folder contains scripts allowing you to run benchmarking tests on Lachesis. It includes the following:

1. `start.sh`: Starts up `N` number of nodes (default=5). This parameter can be changed on line 9: `N=5`. Nodes are
connected as a ring. The `datadir` is created at `/tmp/lachesis-demo`, where a separate data folder is created per node.

2. `stop.sh`: stops all Lachesis nodes from running. The `/tmp/lachesis-demo` directory with node data is deleted.

The `start_dyn_first.sh` and `start_dyn_second.sh` scripts are used in combination to test dynamic participation:

1. `start_dyn_first.sh`: Starts `N` number of nodes specified at runtime, connected together as a ring

2. `start_dyn_second.sh`: Starts `M number` of nodes, which are then connected together as a ring to the existing `N` number node network created by `start_dyn_first.sh`.

Tx-generators are used to generate simple account balance transactions for the network: The tx-generator files (which create and destroy tx-generators) are as follows:

1.  `./txstorm-start.sh`: Starts the tx-generators, the number of transaction  generated per second are specified on line 37:
`--rate=<number of transactions per second.`

2.  `./txstorm-stop.sh`: Destroys the tx-generators

# Installation

[Build lachesis from source](https://github.com/Fantom-foundation/go-lachesis#building-the-source)

Run `go build -o ./build/tx-storm ./cmd/tx-storm`

You should now have `lachesis` and `tx-storm` under the `go-lachesis/build` folder

# How to run `start.sh` with tx-generators

Under `demo` directory:

1. Run `./start.sh`

2. Wait for the nodes to connect

3. Run `./txstorm-start.sh`

Transactions will be continuously send to the network, until stopped.

# Logs

logs are generated under `go-lachesis/txstorm_logs` for each node

# How to run with dynamic participation and tx-generators

Under `demo` directory:

1. Run `.\start_dyn_first.sh`: This creates the first `N` nodes of the network

2. Wait for the nodes to connect

3. Run `.\start_dyn_second.sh`: This creates `M` nodes and attaches to the network created by `.\start_dyn_first.sh`

4. Wait for the nodes to connect

5. Run `.\txstorm-start.sh`


# Stopping

Under `demo` directory

1. Run `txstorm-stop.sh`: Destroy the tx-generators. Note that this can be restarted at any time afterwards, and the
network will start to receieve transactions again (assuming the network is still running)

2. Run `stop.sh`: Destroys the network, and deletes the `/tmp/lachesis-demo` directory

# Parameters

1. `start.sh`:

`N=X`: Specifies `X` nodes to startup

2. `txstorm-start.sh`:

`--rate=Y`: Specfies `Y` transactions to generate per start_dyn_second

3. Minimum emit interval for event blocks

In `go-lachesis/gossip/config_emitter.go`:

`MinEmitInterval:            Z * time.Millisecond`: Set the minimum time Z it takes to emit an event blocks (in Milliseconds)
