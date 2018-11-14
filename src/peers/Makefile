current_dir = $(shell pwd)

clean:
	rm -rf *.pb.go

proto:
	protoc --proto_path=. --go_out=paths=source_relative:. *.proto

build:
	exit 0