.PHONY: all
all: x1

ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

GOPROXY ?= "https://proxy.golang.org,direct"
.PHONY: x1
x1:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w -X github.com/Fantom-foundation/go-opera/cmd/opera/launcher.gitCommit=$${GIT_COMMIT} -X github.com/Fantom-foundation/go-opera/cmd/opera/launcher.gitDate=$${GIT_DATE}" \
	    -o build/x1 \
	    ./cmd/opera

install:
	system/x1-pre-install.sh

	install -d $(DESTDIR)$(PREFIX)/lib
	install -m 0777 build/x1 $(DESTDIR)$(PREFIX)/bin

ifneq ("$(wildcard $(/etc/systemd/system))","")
		install -m 644 system/lib/systemd/system/x1.service $(DESTDIR)/lib/systemd/system
endif

	install -d $(DESTDIR)$(PREFIX)/share/x1/configs/testnet
	install -m 644 system/usr/share/x1/configs/testnet/full-node.toml $(DESTDIR)$(PREFIX)/share/x1/configs/testnet/full-node.toml
	install -m 644 system/usr/share/x1/configs/testnet/api-node.toml $(DESTDIR)$(PREFIX)/share/x1/configs/testnet/api-node.toml
	install -m 644 system/usr/share/x1/configs/testnet/archive-node.toml $(DESTDIR)$(PREFIX)/share/x1/configs/testnet/archive-node.toml

	system/x1-post-install.sh

TAG ?= "latest"
.PHONY: x1-image
x1-image:
	docker build \
    	    --network=host \
    	    -f ./docker/Dockerfile.opera -t "x1:$(TAG)" .

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -coverprofile=cover.prof $$(go list ./... | grep -v '/gossip/contract/' | grep -v '/gossip/emitter/mock' | xargs)
	go tool cover -func cover.prof | grep -e "^total:"

.PHONY: fuzz
fuzz:
	CGO_ENABLED=1 \
	mkdir -p ./fuzzing && \
	go run github.com/dvyukov/go-fuzz/go-fuzz-build -o=./fuzzing/gossip-fuzz.zip ./gossip && \
	go run github.com/dvyukov/go-fuzz/go-fuzz -workdir=./fuzzing -bin=./fuzzing/gossip-fuzz.zip


.PHONY: clean
clean:
	rm -fr ./build/*
