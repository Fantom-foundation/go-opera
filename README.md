# Opera 

EVM-compatible chain secured by the Lachesis consensus algorithm.

## Opera transaction tracing node

Node with implemented OpenEthereum(Parity) transaction tracing API.

```shell
make opera
```
The build output is ```build/opera``` executable.

## Running `opera`

Going through all the possible command line flags is out of scope here,
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `opera` instance.

### Launching a network

You will need a genesis file to join a network, which may be found in https://github.com/Fantom-foundation/lachesis_launch

Launching `opera` readonly (non-validator) node for network specified by the genesis file:

```shell
$ opera --genesis file.g
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

#### Validator

New validator private key may be created with `opera validator new` command.

To launch a validator, you have to use `--validator.id` and `--validator.pubkey` flags to enable events emitter.

```shell
$ opera --nousb --validator.id YOUR_ID --validator.pubkey 0xYOUR_PUBKEY
```

`opera` will prompt you for a password to decrypt your validator private key. Optionally, you can
specify password with a file using `--validator.password` flag.

This branch is for a transaction tracing node. It's not recommended to use this branch for an Opera validator.

### Available methods

```shell
trace_get
trace_transaction
trace_block
trace_filter
```

### Building the source

Building `opera` requires a Go (version 1.15 or later) and a C compiler. Once the dependencies are installed, run:

```shell
make opera
```
The build output is ```build/opera``` executable.

### Running `go-opera` node with tracing ability

It's recommended to launch new node from scratch with CLI option `--tracenode` as this flag has to be set to use stored transaction traces.

```shell
$ opera --genesis /path/to/genesis.g --tracenode
```

### Tracing pre-genesis blocks
If you want to use tracing for pre-genesis blocks, you'll need to import evm history. You can find instructions here: [importing evm history](https://github.com/Fantom-foundation/lachesis_launch/blob/master/docs/import-evm-history.sh)

Then enable JSON-RPC API with option `trace` (`--http.api=eth,web3,net,txpool,ftm,sfc,trace"`)

After that, you have to connect your nodes. Either connect them statically or specify a bootnode:
```shell
$ opera --fakenet 1/5 --bootnodes "enode://verylonghex@1.2.3.4:5050"
```

### Running the demo

For the testing purposes, the full demo may be launched using:
```shell
cd demo/
./start.sh # start the Opera processes
./stop.sh # stop the demo
./clean.sh # erase the chain data
```
Check README.md in the demo directory for more information.
