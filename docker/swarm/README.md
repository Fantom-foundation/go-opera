# Deploy to docker-swarm 


## Configure docker client

1. Set "*_HOST" environments in `set_env.sh`.

2. Put docker-swarm certs into `./ssl`.

3. Login to docker private registry `./10.registry-login.sh` with password.


## Get docker image

1. Build images `cd .. && make`.

2. Upload images to private registry `./20.push-docker-image.sh`.


## Deploy lachesis

1. Create node1-node4 services `./30.create-nodes.sh`.

2. Upgrade node services `./40.upgrade-nodes.sh`.

3. Delete node services `./80.delete-nodes.sh`.


## Node console (example)

1. Get node2 token `export enode2=$(./50.node-console.sh 2 --exec 'admin.nodeInfo.enode')`.

2. Add peer to node3 `./50.node-console.sh 3 --exec "admin.addPeer($enode2)"`.


## Node logs

1. `./swarm service logs node2`.


## Performance testing

use `txn-storm` service to generate transaction streams:

  - start: `./32.txn-storm.sh`;
  - stop: `./82.delete-txn-storm.sh`;

and Prometheus to collect metrics.
