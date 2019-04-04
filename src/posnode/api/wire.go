package api

// Install before go generate:
//  wget https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip
//  unzip protoc-3.6.1-linux-x86_64.zip -x readme.txt -d /usr/local/
//  go get -u github.com/golang/protobuf/protoc-gen-go

//go:generate protoc -I=../../../../../.. -I=. --go_out=plugins=grpc:./ service.proto stored.proto

// NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=api -source=service.pb.go -destination=mock.go NodeServer

import (
	"context"
	"math"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

// StartService starts and returns gRPC server.
func StartService(bind string, svc NodeServer, log func(string, ...interface{}), fake bool) (*grpc.Server, string) {
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	RegisterNodeServer(server, svc)

	var listener net.Listener
	if !fake {
		listener = network.TcpListener(bind)
	} else {
		listener = network.FakeListener(bind)
	}

	log("service start at %v", listener.Addr())
	go func() {
		if err := server.Serve(listener); err != nil {
			log("service stop (%v)", err)
		}
	}()

	return server, listener.Addr().String()
}

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
