all: build



.PHONY: build start stop

build:
	cd ../../.. && docker build -f src/poslachesis/docker/Dockerfile -t "pos-lachesis" .

start:
	./start.sh

stop:
	./stop.sh



.PHONY: blockade

blockade:
	docker build -f Dockerfile.blockade -t "pos-lachesis-blockade" .

blockade-on:
	blockade up

blockade-off:
	blockade destroy

