# X1 

X1 is a simple, fast, and secure EVM-compatible network for the next generation of decentralized applications powered by Lachesis consensus algorithm.

## Explore the Network (Testnet)

- [X1 Explorer](https://explorer.x1-testnet.xen.network)

RPC Endpoints

- https://x1-testnet.xen.network/
- wss://x1-testnet.xen.network

## Run Full Node (Testnet)

You will need a genesis file to join a network, which may be found in https://x1-testnet-genesis.s3.amazonaws.com/x1-testnet.g
You will also need a list of boot nodes to connect to the network. You can find a list of boot nodes in https://x1-testnet-genesis.s3.amazonaws.com/bootnodes.txt.

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

# Download the genesis file and bootnodes list
cd ~
wget https://x1-testnet-genesis.s3.amazonaws.com/x1-testnet.g
wget https://x1-testnet-genesis.s3.amazonaws.com/bootnodes.txt

# Run the node
x1 --genesis x1-testnet.g --genesis.allowExperimental --bootnodes $(paste -sd, bootnodes.txt)
```
