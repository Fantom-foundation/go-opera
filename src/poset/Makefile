.PHONY: build clean proto test

clean:
	rm -rf *.pb.go

proto:
	$(PROTOC) --proto_path=. --proto_path="$(GOPATH)/src" --go_out=paths=source_relative:. *.proto

build:

test:
	go test -race -timeout 45s
