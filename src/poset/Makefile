export GO?=go
export PROTOC?=protoc
export RM?=rm

.PHONY: build clean proto test

clean:
	$(RM) -rf *.pb.go

proto:
	$(PROTOC) --proto_path=. --proto_path="$(GOPATH)/src" --go_out=paths=source_relative:. *.proto

build:

test:
	$(GO) test -tags test -race -cover -timeout 45s
