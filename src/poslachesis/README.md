# PoS-Lachesis node

Package assembles functionality of [network node](../posnode/) and [consensus](../posposet/) into solid Lachesis-node.


### Executables

[`cmd/`](./cmd/) - contains cli (only for fakenet now):

  - run single node: `go run main.go`;
  - for args see `go run main.go --help`;
  - build: `go build .`;


### Docker

[`docker/`](./docker/) - contains docker scripts to try lachesis fakenet:

  - build node docker image "pos-lachesis": `make`;
  - run network of N nodes: `n=N ./start.sh`;
  - drop network: `./stop.sh`;
