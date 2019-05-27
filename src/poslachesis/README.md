# PoS-Lachesis node

Package assembles functionality of [network node](../posnode/) and [consensus](../posposet/) into solid Lachesis-node.

## Executables

[`cli/`](./cli/) - contains cli (only for fakenet now):

  - build: `go build -o lachesis .`;
  - for help use: `./lachesis -h`;
  - run single node: `./lachesis start`;

## Docker

[`docker/`](./docker/) - contains docker scripts to try lachesis fakenet:

  - build node docker image "pos-lachesis": `make`;
  - run network of N nodes: `n=N ./start.sh`;
  - drop network: `./stop.sh`;

## Transfer example.

In this section will be described example of transfering tokens from one node to another.

1. Check balance of our node to check that we have something to transfer.

```sh
./lachesis balance
```

Outputs balance of the node.

```
balance of 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf == 1000000000
```

2. Transfer some amount to another node using `transfer` command

```sh
./lachesis transfer --amount=10000 --receiver=0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0
```

Command returns hex of the transaction we have made.

```
0xab92d4dff02c4dc9f52c298a6cef1e7b73c46215c2bb621d6decfe57c1c0af4b
```

3. Check transaction status `info` command and previous hex of the transfer.

```sh
./lachesis info 0xab92d4dff02c4dc9f52c298a6cef1e7b73c46215c2bb621d6decfe57c1c0af4b
```

Will output transaction information and its status. 

```
transfer 10000 from 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf to 0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0 confirmed
```

4. As soon as transaction status changed to `confirmed` this will result in the balance change.

```
./lachesis balance --peer 0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0
```

Outputs the balance of the peer we've sended the money. Note that while transaction status is uncofirmed there won't be any changes in balance.

```
balance of 0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0 == 10000
```
