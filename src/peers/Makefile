export GO?=go
export PROTOC?=protoc
export RM?=rm

PROTOS:=$(wildcard *proto)
PBGOS:=$(PROTOS:.proto=.pb.go)

.PHONY: build clean proto test buildtests

clean:
	$(RM) -rf *.pb.go

proto: $(PBGOS)

build:

buildtests:

test:
	$(GO) test -race -cover -timeout 45s

%.pb.go: %.proto
	$(PROTOC) --proto_path=. --proto_path="$(GOPATH)/src" --go_out=paths=source_relative:. $<
