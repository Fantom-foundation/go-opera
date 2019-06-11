export GO?=go
export PROTOC?=protoc
export RM?=rm

PROTOS:=$(wildcard *proto)
PBGOS:=$(PROTOS:.proto=.pb.go)
INTERWIRE:=../inter/wire

.PHONY: build clean proto test buildtests

clean:
	$(RM) -rf *.pb.go

proto: $(PROTOS)
	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=paths=source_relative:. $(PROTOS)

build:

buildtests:

test:
	$(GO) test -tags test -count=1 -race -cover -timeout 45s

