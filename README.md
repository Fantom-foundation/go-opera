# Opera transaction tracing node

Node with implemented OpenEthereum(Parity) transaction tracing API

## Available methods

```shell
trace_transaction
trace_block
trace_filter
```

## Running `go-opera` node with tracing ability

It's recomended to install new node from scratch with CLI option `--tracenode` and this flag has to be set to use stored transaction traces.

When you want to use tracing for pre-genesis blocks, then you need to import evm history. You can find instructions here: [importing evm history](https://github.com/uprendis/lachesis_launch/blob/release/opera-v1.0.0-rc1-mainnet/docs/import-evm-history.sh)

Then enable also JSON-RPC API with option `trace` (`--rpcapi=eth,web3,net,txpool,ftm,sfc,trace"`)