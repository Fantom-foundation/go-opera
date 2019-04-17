.PHONY: build start stop

all: build


build:
	cd ../../.. && docker build -f src/poslachesis/docker/Dockerfile -t "pos-lachesis" .

start:
	./start.sh

stop:
	./stop.sh