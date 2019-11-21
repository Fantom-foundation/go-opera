# Deploy to docker-swarm 


## Configure docker client

1. Set "*_HOST" environments in `_params.sh`.

2. Put docker-swarm certs into `./ssl`.

3. Login to docker private registry `./10.registry-login.sh` with password.


## Get docker image

1. Build images `cd .. && make` (use GOPROXY and TAG env vars optionally).

2. Upload images to private registry `./20.push-docker-image.sh` (use TAG env var optionally).


## Deploy lachesis

1. Create node0-node2 services `./30.create-nodes.sh` (use N and TAG env vars optionally).

2. Delete node services `./80.delete-nodes.sh`.


## Node console (example)

1. Get node1 token `export enode=$(./50.node-console.sh 1 --exec 'admin.nodeInfo.enode')`.

2. Add peer to node2 `./50.node-console.sh 2 --exec "admin.addPeer($enode)"`.


## Node logs

1. `./swarm service logs node2`.


## Performance testing

use `tx-storm` service to generate transaction streams:

  - start: `./31.txstorm-on.sh`  (use N and TAG env vars optionally);
  - stop: `./81.txstorm-off.sh`;

and Prometheus to collect metrics:

  - start: `./32.prometheus-on.sh`;
  - stop: `./82.prometheus-off.sh`;
