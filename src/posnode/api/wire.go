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
	"encoding/base64"
	"errors"
	"math"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

// StartService starts and returns gRPC server.
func StartService(bind, serverID string, serverKey *common.PrivateKey, svc NodeServer, log func(string, ...interface{}), listen network.ListenFunc) (*grpc.Server, string) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(serverInterceptor(serverID, serverKey)),
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

// SignGRPCData before send it
func SignGRPCData(data interface{}, key *common.PrivateKey) (sign, pubKey string) {
	b, _ := data.([]byte)
	R, S, _ := key.Sign(b)

	sign = crypto.EncodeSignature(R, S)
	pubKey = base64.StdEncoding.EncodeToString(key.Public().Bytes())

	return
}

// GetInfoFromMetadata get info from request/response metadata
func GetInfoFromMetadata(md metadata.MD) (rID, rSign, rPub string, err error) {
	// Get ID from metadata
	id, ok := md["id"]
	if !ok {
		err = errors.New("id not found")
		return
	}

	// Get signature from metadata
	sign, ok := md["sign"]
	if !ok {
		err = errors.New("sign not found")
		return
	}

	// Get public key from metadata
	pub, ok := md["pub"]
	if !ok {
		err = errors.New("pub not found")
		return
	}

	if id[0] == "" || sign[0] == "" || pub[0] == "" {
		err = errors.New("Empty data")
		return
	}

	rID = id[0]
	rSign = sign[0]
	rPub = pub[0]

	return
}

// CheckSignData from response / request
func CheckSignData(req interface{}, sign string, key *common.PublicKey) (bool, error) {
	hash, _ := req.([]byte)
	r, s, err := crypto.DecodeSignature(sign)
	if err != nil {
		return false, status.Errorf(codes.Internal, "Cannot decode signature")
	}

	return key.Verify(hash, r, s), nil
}

// ValidateID from response / request
func ValidateID(id string, key *common.PublicKey) bool {
	p := hash.PeerOfPubkey(key)
	return p.Hex() == id
}

func serverInterceptor(serverID string, serverKey *common.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get metadata from context
		id, sign, pub, err := getMetadata(ctx)
		if err != nil {
			return nil, err
		}

		// Extract common.Pubkey from string
		key, err := common.StringToPubkey(pub)
		if err != nil {
			return nil, err
		}

		// Validation ID
		if ok := ValidateID(id, key); !ok {
			return nil, status.Errorf(codes.Internal, "ID is invalid")
		}

		// Check signature of request
		isValid, err := CheckSignData(req, sign, key)
		if err != nil {
			return nil, err
		}

		if !isValid {
			return nil, status.Errorf(codes.Internal, "Signature is invalid")
		}

		// Process request
		resp, err := handler(ctx, req)

		// Sign response
		sign, pubKey := SignGRPCData(resp, serverKey)

		// Create new metadata for current response
		md := metadata.Pairs("id", serverID, "sign", sign, "pub", pubKey)
		grpc.SetTrailer(ctx, md)

		return resp, err
	}
}

func getMetadata(ctx context.Context) (rID, rSign, rPub string, err error) {
	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
		return
	}

	rID, rSign, rPub, err = GetInfoFromMetadata(md)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, err.Error())
		return
	}

	return
}
