TAG ?= "latest"

.PHONY: all
all: opera

.PHONY: opera
opera:
	docker build \
    	    --network=host \
    	    --build-arg GOPROXY=$(GOPROXY) \
    	    -f ./Dockerfile.opera -t "opera:$(TAG)" ..
