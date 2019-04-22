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
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

// StartService starts and returns gRPC server.
func StartService(bind string, svc NodeServer, log func(string, ...interface{}), listen network.ListenFunc) (*grpc.Server, string) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(serverInterceptor),
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

func serverInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Get metadata from context
	_, sign, pub, err := getMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Extract common.Pubkey from string
	key, err := extractPubkey(pub)
	if err != nil {
		return nil, err
	}

	// Check signature of request
	isInvalid, err := checkSignRequest(req, sign, key)
	if err != nil {
		return nil, err
	}

	if !isInvalid {
		return nil, status.Errorf(codes.Internal, "Signature is invalid")
	}

	return handler(ctx, req)
}

func getMetadata(ctx context.Context) (rID, rSign, rPub string, err error) {
	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
		return
	}

	// Get ID from metadata
	// Currently we check only exist status ID and don't use it as is.
	id, ok := md["client_id"]
	if !ok {
		err = status.Errorf(codes.InvalidArgument, "client_id not found")
		return
	}

	// Get signature from metadata
	sign, ok := md["client_sign"]
	if !ok {
		err = status.Errorf(codes.InvalidArgument, "client_sign not found")
		return
	}

	// Get public key from metadata
	pub, ok := md["client_pub"]
	if !ok {
		err = status.Errorf(codes.InvalidArgument, "client_pub not found")
		return
	}

	if id[0] == "" || sign[0] == "" || pub[0] == "" {
		err = status.Errorf(codes.InvalidArgument, "Empty data")
		return
	}

	rID = id[0]
	rSign = sign[0]
	rPub = pub[0]

	return
}

func extractPubkey(pub string) (*common.PublicKey, error) {
	var bb []byte
	for _, ps := range strings.Split(strings.Trim(pub, "[]"), " ") {
		pi, _ := strconv.Atoi(ps)
		bb = append(bb, byte(pi))
	}

	key := common.BytesToPubkey(bb)
	if key == nil {
		return nil, status.Errorf(codes.Internal, "Pubkey is invalid")
	}

	return key, nil
}

func checkSignRequest(req interface{}, sign string, key *common.PublicKey) (bool, error) {
	hash, _ := req.([]byte)
	r, s, err := crypto.DecodeSignature(sign)
	if err != nil {
		return false, status.Errorf(codes.Internal, "Cannot decode signature")
	}

	return key.Verify(hash, r, s), nil
}
