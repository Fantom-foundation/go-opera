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

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

// StartService starts and returns gRPC server.
func StartService(bind string, key *common.PrivateKey, svc NodeServer, log func(string, ...interface{}), listen network.ListenFunc) (
	*grpc.Server, string) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(ServerAuth(key)),
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	RegisterNodeServer(server, svc)

	listener := listen(bind)

	log("service start at %v", listener.Addr())
	go func() {
		if err := server.Serve(listener); err != nil {
			log("service stop (%v)", err)
		}
	}()

	return server, listener.Addr().String()
}

// GrpcPeerHost extracts client's host from grpc context.
func GrpcPeerHost(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		panic("GrpcPeerHost should be called from gRPC handler only")
	}

	addr := p.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	return host
}

// GrpcPeerID extracts client's ID from grpc context.
func GrpcPeerID(ctx context.Context) hash.Peer {
	id, ok := ctx.Value(peerID{}).(hash.Peer)
	if !ok {
		panic("GrpcPeerID should be called from gRPC handler only")
	}

	return id
}

// ServerPeerID makes context for gRPC call to get server-peer id.
func ServerPeerID(parent context.Context) (id *hash.Peer, ctx context.Context) {
	if parent == nil {
		parent = context.Background()
	}

	id = &hash.Peer{}
	ctx = context.WithValue(parent, peerID{}, func(x hash.Peer) {
		*id = x
	})

	return
}
