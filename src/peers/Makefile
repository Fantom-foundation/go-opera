export GO?=go
export PROTOC?=protoc
export RM?=rm

.PHONY: build clean proto test

clean:
	$(RM) -rf *.pb.go

proto:
	$(PROTOC) --proto_path=. --go_out=paths=source_relative:. *.proto

build:

test:
	$(GO) test -race -cover -timeout 45s
