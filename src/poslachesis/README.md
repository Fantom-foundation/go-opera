# PoS-Lachesis node

Package assembles functionality of [network node](../posnode/) and [consensus](../posposet/) into solid Lachesis-node.

## Executables

[`cli/`](./cli/) - contains cli (only for fakenet now):

  - build: `go build -o lachesis .`;
  - for help use: `./lachesis -h`;
  - run single node: `./lachesis start`;
  - get id of the node: `./lachesis id`;
  - get balance of the node: `./lachesis balance`;
  - transfer balance: `./lachesis transfer --amount=1000 --receiver=nodeId`;
	- Note: all pending transaction which is not yet put in an event block, will be displayed at `./lachesis balance` call;

## Docker

[`docker/`](./docker/) - contains scripts to try lachesis fakenet with Docker:


### for common purpose:

  - build node docker image "pos-lachesis": `make build`;
  - run network of N nodes: `n=N ./start.sh`;
  - drop network: `./stop.sh`;


### for testing network failures:

  - build node docker image "pos-lachesis-blockade": `make build blockade`;
  - install blockade and add it to "$PATH": `pip install blockade`;
  - make blockade.yaml config: `n=N ./blockade.sh`;
  - test lachesis network with blockade [`commands`](https://github.com/worstcase/blockade/blob/master/docs/commands.rst);
