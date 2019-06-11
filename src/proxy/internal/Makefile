export PROTOC?=protoc
export RM?=rm

PROTOS:=$(wildcard *proto)
PBGOS:=$(PROTOS:.proto=.pb.go)
INTERWIRE:=../../inter/wire

.PHONY: build clean proto

clean:
	$(RM) -rf *.pb.go

proto: $(PROTOS)
	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=plugins=grpc,import_path=internal:. $(PROTOS)

build:

#%.pb.go: %.proto
#	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=plugins=grpc,import_path=internal:. $<
#
#app.pb.go: app.proto
#	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=plugins=grpc:. $<
#
#ctrl.pb.go: ctrl.proto app.pb.go
#	$(PROTOC) --proto_path=. -I "$(INTERWIRE)" --proto_path="$(GOPATH)/src" --go_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/golang/protobuf/ptypes/empty:. $<
