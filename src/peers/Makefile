.PHONY: build clean proto test

clean:
	rm -rf *.pb.go

proto:
	protoc --proto_path=. --go_out=paths=source_relative:. *.proto

build:

test:
	go test -race -timeout 45s
