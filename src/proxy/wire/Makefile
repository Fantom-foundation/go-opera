export PROTOC?=protoc
export RM?=rm

.PHONY: build clean proto

clean:
	$(RM) -rf *.pb.go

proto:
	$(PROTOC) --go_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/golang/protobuf/ptypes/empty,import_path=wire:. *.proto

build:
