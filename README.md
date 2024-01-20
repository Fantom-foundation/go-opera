# X1 Blockchain

X1 is a simple, fast, and secure EVM-compatible network for the next generation of decentralized applications powered by Lachesis consensus algorithm.

Chain ID: `204005`

## Explore the Network (Testnet)

- [X1 Explorer](https://explorer.x1-testnet.xen.network)

### RPC Endpoints

- https://x1-testnet.xen.network/
- wss://x1-testnet.xen.network

## Run Full Node (Testnet)

> Quick start a full node and run in the foreground

```shell
# Install dependencies (ex: ubuntu)
apt update -y
apt install -y golang wget git make

# Clone and build the X1 binary
git clone --branch x1 https://github.com/FairCrypto/go-x1
cd go-x1
make x1
cp build/x1 /usr/local/bin

# Run the node
x1 --testnet --syncmode snap
```

### Examples

> Run with Xenblocks reporting enabled
```shell
x1 --testnet --syncmode snap --xenblocks-endpoint ws://xenblocks.io:6668
```

## Run an API Node (Testnet)

An API node corresponds to a full node that has the RPC server activated. 
To set up an API node, follow the complete node instructions mentioned earlier, 
but execute the process with the --http and/or --ws flags.

### Examples

> Run with RPC server enabled
```shell
x1 --testnet --http --http.port 8545 --ws --ws.port 8546
```

> Run with RPC server open to the world (only run if you know what you are doing)
```shell
x1 --testnet --http --http.port 8545 --http.addr 0.0.0.0 --http.vhosts "*" --http.corsdomain "*" --ws --ws.addr 0.0.0.0 --ws.port 8546 --ws.origins "*"
```

## Run an Archive Node (Testnet)

Synchronizing an archive node is a time-consuming process with substantial disk space requirements. 
However, it provides the capability to retrieve historical data through queries.

```shell
x1 --testnet --gcmode archive --syncmode full
```

## Run a Validator Node (Testnet)

See [here](docs/validators) for more information on running a validator node.
