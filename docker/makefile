BABBLE_VERSION=0.3.0
DUMMY_VERSION=0.3.0
WATCHER_VERSION=0.0.2
GLIDER_VERSION=0.0.2

all: lachesis-image dummy-image watcher-image glider-image

demo: lachesis-image dummy-image watcher-image

lachesis-image:
	CGO_ENABLED=0 go build \
		--ldflags "-X github.com/andrecronje/lachesis/src/version.GitCommit=`git rev-parse HEAD`" \
		-o lachesis/lachesis ../cmd/lachesis/
	docker build --no-cache=true -t andrecronje/lachesis:$(BABBLE_VERSION) lachesis/

dummy-image:
	CGO_ENABLED=0 go build \
		-o dummy/dummy ../cmd/dummy_client/
	docker build --no-cache=true -t andrecronje/dummy:$(DUMMY_VERSION) dummy/

watcher-image:
	docker build --no-cache=true -t andrecronje/watcher:$(WATCHER_VERSION) -t andrecronje/watcher:latest watcher/

glider-image:
	docker build --no-cache=true -t andrecronje/glider:$(GLIDER_VERSION) glider/

.PHONY: all demo lachesis-image dummy-image watcher-image glider-image
