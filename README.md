# Opera 

EVM-compatible chain secured by the Lachesis consensus algorithm.

## Building the source

Building `opera` requires both a Go (version 1.14 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run

```shell
make opera
```
The build output is ```build/opera``` executable.

Do not clone the project into $GOPATH, due to the Go Modules. Instead, use any other location.

## Running `opera`

Going through all the possible command line flags is out of scope here,
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `opera` instance.

### Mainnet

Launching `opera` for mainnet with default settings:

```shell
$ opera
```

### Configuration

As an alternative to passing the numerous flags to the `opera` binary, you can also pass a
configuration file via:

```shell
$ opera --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to
export your existing configuration:

```shell
$ opera --your-favourite-flags dumpconfig
```

#### Docker quick start

One of the quickest ways to get Lachesis up and running on your machine is by using
Docker:

```shell
cd docker/
make
docker run -d --name opera-node -v /home/alice/opera:/root \
           -p 5050:5050 \
          "opera" \
          --port 5050 \
          --nat extip:YOUR_IP
```

This will start `opera` with ```--port 5050 --nat extip:YOUR_IP``` arguments, with DB files inside ```/home/alice/opera/.opera```

Do not forget `--http.addr 0.0.0.0`, if you plan to access RPC from other containers
and/or hosts. By default, `opera` binds to the local interface and RPC endpoints is not
accessible from the outside.

To find out your enode ID, use:
```shell
docker exec -i opera-node /opera --exec "admin.nodeInfo.enode" attach
```
To get the logs:
```
docker logs opera-node
```

#### Validator

To launch a validator, you have to use `--validator` flag to enable events emitter. Also you have to either use `--unlock` / `--password` flags or unlock
validator account manually. Validator account should be unlocked for signing events.

```shell
$ opera --nousb --validator 0xADDRESS --unlock 0xADDRESS --password /path/to/password
```

#### Participation in discovery

Optionally you can specify your public IP to straighten connectivity of the network.
Ensure your TCP/UDP p2p port (5050 by default) isn't blocked by your firewall.

```shell
$ opera --nat extip:1.2.3.4
```

## Dev

### Running testnet

To run a testnet node, you have to add `--testnet` flag every time you use `opera`:

```shell
$ opera --testnet # launch node
$ opera --testnet attach # attach to IPC
$ opera --testnet account new # create new account
```

### Testing

Lachesis has extensive unit-testing. Use the Go tool to run tests:
```shell
go test ./...
```

If everything goes well, it should output something along these lines:
```
ok  	github.com/Fantom-foundation/go-opera/app	0.033s
?   	github.com/Fantom-foundation/go-opera/cmd/cmdtest	[no test files]
ok  	github.com/Fantom-foundation/go-opera/cmd/opera	13.890s
?   	github.com/Fantom-foundation/go-opera/cmd/opera/metrics	[no test files]
?   	github.com/Fantom-foundation/go-opera/cmd/opera/tracing	[no test files]
?   	github.com/Fantom-foundation/go-opera/crypto	[no test files]
?   	github.com/Fantom-foundation/go-opera/debug	[no test files]
?   	github.com/Fantom-foundation/go-opera/ethapi	[no test files]
?   	github.com/Fantom-foundation/go-opera/eventcheck	[no test files]
?   	github.com/Fantom-foundation/go-opera/eventcheck/basiccheck	[no test files]
?   	github.com/Fantom-foundation/go-opera/eventcheck/gaspowercheck	[no test files]
?   	github.com/Fantom-foundation/go-opera/eventcheck/heavycheck	[no test files]
?   	github.com/Fantom-foundation/go-opera/eventcheck/parentscheck	[no test files]
ok  	github.com/Fantom-foundation/go-opera/evmcore	6.322s
?   	github.com/Fantom-foundation/go-opera/gossip	[no test files]
?   	github.com/Fantom-foundation/go-opera/gossip/emitter	[no test files]
ok  	github.com/Fantom-foundation/go-opera/gossip/filters	1.250s
?   	github.com/Fantom-foundation/go-opera/gossip/gasprice	[no test files]
?   	github.com/Fantom-foundation/go-opera/gossip/occuredtxs	[no test files]
?   	github.com/Fantom-foundation/go-opera/gossip/piecefunc	[no test files]
ok  	github.com/Fantom-foundation/go-opera/integration	21.640s
```

### Operating a private network (fakenet)

Fakenet is a private network optimized for your private testing.
It'll generate a genesis containing N validators with equal stakes.
To launch a validator in this network, all you need to do is specify a validator ID you're willing to launch.

Pay attention that validator's private keys are deterministically generated in this network, so you must use it only for private testing.

Maintaining your own private network is more involved as a lot of configurations taken for
granted in the official networks need to be manually set up.

To run the fakenet with just one validator (which will work practically as a PoA blockchain), use:
```shell
$ opera --fakenet 1/1
```

To run the fakenet with 5 validators, run the command for each validator:
```shell
$ opera --fakenet 1/5 # first node, use 2/5 for second node
```

If you have to launch a non-validator node in fakenet, use 0 as ID:
```shell
$ opera --fakenet 0/5
```

After that, you have to connect your nodes. Either connect them statically or specify a bootnode:
```shell
$ opera --fakenet 1/5 --bootnodes "enode://verylonghex@1.2.3.4:5050"
```

### Running the demo

For the testing purposes, the full demo may be launched using:
```shell
make
cd demo/
./start.sh # start the Opera processes
./stop.sh # stop the demo and erase the chain data
```
