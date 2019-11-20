all: lachesis tx-storm

TAG ?= "latest"

.PHONY: lachesis
lachesis:
	cd ../ && \
	docker build \
	    --network=host \
	    --build-arg GOPROXY=$(GOPROXY) \
	    -f docker/Dockerfile.lachesis -t "lachesis:$(TAG)" .


.PHONY: tx-storm
tx-storm:
	cd ../ && \
	docker build \
	    --network=host \
	    --build-arg GOPROXY=$(GOPROXY) \
	    -f docker/Dockerfile.tx-storm -t "tx-storm:$(TAG)" .
