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

[`docker/`](./docker/) - contains docker scripts to try lachesis fakenet:

  - build node docker image "pos-lachesis": `make`;
  - run network of N nodes: `n=N ./start.sh`;
  - drop network: `./stop.sh`;

## Sentry

If you want to use Sentry for catch error, please use this [`link`](https://github.com/getsentry/onpremise).
And include `--dsn` key for `./docker/start.sh`

Example: `--dsn="http://64f6a4a7aaba4aa0a12fedd4d8f7aa61@localhost:9000/1"`

Where: `--dsn="http://<sentry public key>@<host>:<port>/<project id>"`
