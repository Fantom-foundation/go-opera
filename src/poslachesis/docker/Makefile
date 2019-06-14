all: build


.PHONY: build
build:
	cd ../../.. && docker build -f src/poslachesis/docker/Dockerfile -t "pos-lachesis" .


.PHONY: blockade
blockade:
	docker build -f Dockerfile.blockade -t "pos-lachesis-blockade" .
