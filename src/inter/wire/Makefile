export PROTOC?=protoc
export RM?=rm

PROTOS:=$(wildcard *proto)
PBGOS:=$(PROTOS:.proto=.pb.go)

.PHONY: build clean proto

clean:
	$(RM) -rf *.pb.go

proto: $(PBGOS)

build:

%.pb.go: %.proto
	$(PROTOC) --proto_path=. --proto_path="$(GOPATH)/src" --go_out=plugins=grpc,import_path=internal:. $<

wire.pb.go: wire.proto
