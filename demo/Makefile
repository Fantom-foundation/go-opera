export DOCKER?=docker
export RM?=rm

.PHONY: up build conf start watch stop demo bombard

nodes = 4
txs = 10

up: conf start watch

# build lachesis and dummy docker images for local registry
build:
	$(DOCKER) build -f ./Dockerfile.dummy -t go-dummy ../
	$(DOCKER) build -f ./Dockerfile.node -t go-lachesis ../

# generate keys and peers.json file
conf: 
	$(RM) -rf conf
	./scripts/build-conf.sh $(nodes)

# launch testnet
start:
	./scripts/run-testnet.sh $(nodes)

# monitor stats
watch: 
	./scripts/watch.sh $(nodes)

# stop testnet
stop: 
	./scripts/stop-testnet.sh

# open terminals to use the dummy chat application
demo: 
	./scripts/demo.sh $(nodes)

# send a bunch of transactions
bombard: 
	./scripts/bombard.sh $(nodes) $(txs)
