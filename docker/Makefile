all: lachesis txn-storm


.PHONY: lachesis
lachesis:
	cd ../ && \
	docker build \
	    --network=host \
	    --build-arg GOPROXY=$(GOPROXY) \
	    -f docker/Dockerfile.lachesis -t "lachesis" .


.PHONY: txn-storm
txn-storm:
	cd ../ && \
	docker build \
	    --network=host \
	    --build-arg GOPROXY=$(GOPROXY) \
	    -f docker/Dockerfile.txn-storm -t "txn-storm" .
