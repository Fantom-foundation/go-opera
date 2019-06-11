export PROTOC?=protoc
export RM?=rm

PROTOS:=$(wildcard *proto)
PBGOS:=$(PROTOS:.proto=.pb.go)
INTERWIRE:=../inter/wire

.PHONY: build clean proto buildtests

clean:
	$(RM) -rf *.pb.go

proto: $(PBGOS)

build:

buildtests:

%.pb.go: %.proto
	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=plugins=grpc:. $<

account.pb.go: account.proto
