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

If you want to use Sentry, please use this [`link`](https://github.com/getsentry/onpremise) for setup your own copy.

Don't forget include `--dsn` key to `./docker/start.sh`

Example: `--dsn="http://64f6a4a7aaba4aa0a12fedd4d8f7aa61@localhost:9000/1"`

Where: `--dsn="http://<sentry public key>@<sentry host>:<port>/<project id>"`

### Tips

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