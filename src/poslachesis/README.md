# PoS-Lachesis node

Package assembles functionality of [network node](../posnode/) and [consensus](../posposet/) into solid Lachesis-node.

## Executables

[`cli/`](./cli/) - contains cli (only for fakenet now):

  - build: `go build -o lachesis .`;
  - for help use: `./lachesis -h`;
  - run single node: `./lachesis start`;


### transfer example

* Check the balance to ensure the node have something to transfer
```sh
./lachesis balance
```
 outputs balance of the node.
```
balance of 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf == 1000000000
```

* Transfer some amount to another node using `transfer` command
```sh
./lachesis transfer --amount=10000 --receiver=0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0
```
 returns hex of the outgoing transaction
```
0xab92d4dff02c4dc9f52c298a6cef1e7b73c46215c2bb621d6decfe57c1c0af4b
```

* Check the transaction status using `info` command and previous hex of the transfer
```sh
./lachesis info 0xab92d4dff02c4dc9f52c298a6cef1e7b73c46215c2bb621d6decfe57c1c0af4b
```
 outputs transaction information and it's status
```
transfer 10000 from 0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf to 0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0 confirmed
```

* As soon as transaction status changed to `confirmed` this will result in the balance change of both nodes
```
./lachesis balance --peer 0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0
```
 outputs the balance of the peer we've sent the money. Note that while transaction status is unconfirmed there won't be any changes in balance
```
balance of 0xf2610eb8185d8120e0e71f0d5f1fc74e3b187646a6a0aee169ca242a6b599fc0 == 10000
```


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

## Sentry

If you want to use Sentry, please use this [`link`](https://github.com/getsentry/onpremise) for setup your own copy.

Don't forget include `--dsn` key to `./docker/start.sh`

Example: `--dsn="http://64f6a4a7aaba4aa0a12fedd4d8f7aa61@localhost:9000/1"`

where: `--dsn="http://<sentry public key>@<sentry host>:<port>/<project id>"`

### tips

If you have an error about Sentry connection, check next steps:

  - `docker-compose.yml` should include `SENTRY_SECRET_KEY` which you should generate using link above.
  ```
  environment:
    SENTRY_SECRET_KEY: !!!SECRET_KEY!!!
  ``` 

  - Try to use custom network for `docker-compose.yml`:
  ```
  networks:
    custom_network:
      driver: bridge
      name: lachesis // The name should be the same as for `pos-lachesis` and as for Sentry containers
  ```

  - Don't forget add this network for each service:

  Example:
  ```
  worker:
    <<: *defaults
    command: run worker
    networks:
      custom_network:
  ```

  - If you have empty Client DSN links, add next line to `config.yml`:
  ```
  system.url-prefix: http://<sentry host>:<port>
  ```
