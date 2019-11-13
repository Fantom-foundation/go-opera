all: lachesis tx-storm


.PHONY: lachesis
lachesis:
	cd ../ && \
	docker build \
	    --network=host \
	    --build-arg GOPROXY=$(GOPROXY) \
	    -f docker/Dockerfile.lachesis -t "lachesis" .


.PHONY: tx-storm
tx-storm:
	cd ../ && \
	docker build \
	    --network=host \
	    --build-arg GOPROXY=$(GOPROXY) \
	    -f docker/Dockerfile.tx-storm -t "tx-storm" .
