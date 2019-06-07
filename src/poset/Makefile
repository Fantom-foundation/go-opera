export GO?=go
export PROTOC?=protoc
export RM?=rm

PROTOS:=$(wildcard *proto)
PBGOS:=$(PROTOS:.proto=.pb.go)
INTERWIRE:=../inter/wire

.PHONY: build clean proto test

clean:
	$(RM) -rf *.pb.go

proto: $(PBGOS)

build:

test:
	$(GO) test -tags test -count=1 -race -cover -timeout 45s

%.pb.go: %.proto
	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=paths=source_relative:. $<

frame.pb.go: frame.proto

flagTableWrapper.pb.go: flagTableWrapper.proto

block.pb.go: block.proto

root.pb.go: root.proto

event.pb.go: event.proto

roundInfo.pb.go: roundInfo.proto
