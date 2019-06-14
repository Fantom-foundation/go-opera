# PoS-Lachesis node

Package assembles functionality of [network node](../posnode/) and [consensus](../posposet/) into solid Lachesis-node.


## Executables

[`cli/`](./cli/) - contains cli:

  - build: 
    - in root directory: `glide install`;
    - in src/poslachesis/cli: `go build -o lachesis .`;
  - for help use: `./lachesis -h`;
  - run single node: `./lachesis start`;


## Docker

[`docker/`](./docker/) - contains scripts to try lachesis (only for fakenet now) with Docker:

### for common purpose

  - build node docker image "pos-lachesis": `make build`;
  - run network of x nodes: `N=x ./start.sh`;
  - stop network: `./stop.sh`;

### the same with Prometheus service:

  - set `PROMETHEUS=yes` before `./start.sh`;

### the same with Sentry service

  - set `SENTRY=yes` before `./start.sh`;
  - remove Sentry database: `sentry/clean.sh`;

During first run, you'll get offer to create Sentry-account. Note that account exist only inside local copy Sentry and don't affect to sentry.io.
After start up go to `http://localhost:9000` and sign in using that Sentry-account to see and management logs from all running local nodes.
Logs are grouped and marked with color (info - blue, warn - yellow, error - red).
Each log include: environment info, message about error, code line (in case error).

### for testing network failures

  - build node docker image "pos-lachesis-blockade": `make build blockade`;
  - install blockade and add it to "$PATH": `pip install blockade`;
  - use `./start_blockade.sh` instead normal `./start.sh`;
  - test lachesis network with blockade [`commands`](https://github.com/worstcase/blockade/blob/master/docs/commands.rst);


## Stake transfer example

from [`docker/`](./docker/) dir

* Start network:
```sh
N=3 ./start.sh
```

* Check the stake to ensure the node1 have something to transfer:
```sh
docker exec -ti pos-lachesis-node-1 /lachesis stake
```
 output shows address and stake:
```
stake of 0x322931dd144ae13865196dde4a0a9e8acff53e26a6dec249c595140e343c8054 == 1000000000
```

* Get node2 address:
```sh
docker exec -ti pos-lachesis-node-2 /lachesis id
```
 output shows address:
```
0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf
```

* Transfer some amount from node1 to node2 address as receiver (index should be unique for sender):
```sh
docker exec -ti pos-lachesis-node-1 /lachesis transfer --index 0 --amount=50000 --receiver=0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf
```
 output shows unique hash the outgoing transaction:
```
0x93be31593b40e7665929ab60f3a761f0d206bc5dac1d43d6622382008a595817
```

* Check the transaction status by its unique hash:
```sh
docker exec -ti pos-lachesis-node-1 /lachesis txn 0x93be31593b40e7665929ab60f3a761f0d206bc5dac1d43d6622382008a595817
```
 outputs shows transaction information and status:
```
transfer 50000 to 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf included into event 0x8ab6ab4fb979225d0c21b5a5c914fecdc7e49c28c4dd434bbcc7bb23eb019ea5 and is not confirmed yet
```

* As soon as transaction status changed to `confirmed`:
```
transfer 50000 to 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf included into event 0x8ab6ab4fb979225d0c21b5a5c914fecdc7e49c28c4dd434bbcc7bb23eb019ea5 and is confirmed by block 56
```
 you will see new stake of both nodes:
 
```sh
docker exec -ti pos-lachesis-node-1 /lachesis stake
stake of 0x322931dd144ae13865196dde4a0a9e8acff53e26a6dec249c595140e343c8054 == 999950000

docker exec -ti pos-lachesis-node-2 /lachesis stake
stake of 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf == 1000500000

docker exec -ti pos-lachesis-node-3 /lachesis stake --peer 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf
stake of 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf == 1000050000
```
