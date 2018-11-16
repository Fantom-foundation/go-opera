.PHONY: build clean proto

clean:
	rm -rf *.pb.go

proto:
	protoc --go_out=plugins=grpc,import_path=internal:. *.proto

build:
