.PHONY: build clean proto

clean:
	rm -rf *.pb.go

proto:
	$(PROTOC) --go_out=plugins=grpc,import_path=internal:. *.proto

build:
