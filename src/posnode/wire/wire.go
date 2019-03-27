package wire

// Install before go generate:
//  wget https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip
//  unzip protoc-3.6.1-linux-x86_64.zip -x readme.txt -d /usr/local/
//  go get -u github.com/golang/protobuf/protoc-gen-go

//go:generate protoc --go_out=plugins=grpc:./ event.proto service.proto
// NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=wire -source=service.pb.go -destination=mock.go NodeServer

import (
	"context"
	"net"

	"google.golang.org/grpc/peer"
)

func GrpcPeerHost(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		addr := p.Addr.String()
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			panic(err)
		}
		return host
	}
	panic("gRPC-peer network address is undefined")
}
