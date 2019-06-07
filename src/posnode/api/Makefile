export PROTOC?=protoc
export RM?=rm

PROTOS:=$(wildcard *proto)
PBGOS:=$(PROTOS:.proto=.pb.go)
INTERWIRE:=../../inter/wire

.PHONY: build clean proto

clean:
	$(RM) -rf *.pb.go

proto: $(PBGOS)

build:

%.pb.go: %.proto
	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=plugins=grpc,import_path=internal:. $<

service.pb.go: service.proto

stored.pb.go: stored.proto
