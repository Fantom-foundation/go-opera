# Opera 

EVM-compatible chain secured by the Lachesis consensus algorithm.

## Opera transaction tracing node

Node with implemented OpenEthereum(Parity) transaction tracing API.

Code [release/txtracing/1.0.1-rc.2](https://github.com/Fantom-foundation/go-opera/tree/release/txtracing/1.0.1-rc.2)

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

