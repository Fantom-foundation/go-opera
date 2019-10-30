all: build


.PHONY: build
build:
	cd ../ && \
	docker build \
	    --network=host \
	    --build-arg GOPROXY=$(GOPROXY) \
	    -f docker/Dockerfile -t "lachesis" .
