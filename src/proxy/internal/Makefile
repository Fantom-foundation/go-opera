export PROTOC?=protoc
export RM?=rm

.PHONY: build clean proto

clean:
	$(RM) -rf *.pb.go

proto:
	$(PROTOC) --go_out=plugins=grpc,import_path=internal:. *.proto

build:
