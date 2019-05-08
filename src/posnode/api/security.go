package api

import (
	"context"
	"encoding/base64"
	"errors"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// peerID is a internal key for context.Value().
type peerID struct{}

// ClientAuth makes client-side interceptor for identification.
func ClientAuth(key *common.PrivateKey) grpc.UnaryClientInterceptor {
	pub := key.Public().Base64()

	return func(ctx context.Context, method string, req interface{}, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// request:

		sign := signData(req, key)
		md := metadata.Pairs("sign", sign, "pub", pub)
		ctx = metadata.NewOutgoingContext(ctx, md)

		var answer metadata.MD
		opts = append(opts, grpc.Trailer(&answer))
		err := invoker(ctx, method, req, resp, cc, opts...)
		if err != nil {
			return err
		}

		// response:

		sign, pub, err := readMetadata(answer)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}

		err = verifyData(resp, sign, pub)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}

		if set, ok := ctx.Value(peerID{}).(func(hash.Peer)); ok {
			serverID := hash.PeerOfPubkey(pub)
			set(serverID)
		}

		return nil
	}
}

// ServerAuth makes server-side interceptor for identification.
func ServerAuth(key *common.PrivateKey) grpc.UnaryServerInterceptor {
	pub := base64.StdEncoding.EncodeToString(key.Public().Bytes())

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// request:

		clientSign, clientPub, err := parseContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		err = verifyData(req, clientSign, clientPub)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		clientID := hash.PeerOfPubkey(clientPub)
		ctx = context.WithValue(ctx, peerID{}, clientID)

		// response:

		resp, err := handler(ctx, req)

		sign := signData(resp, key)
		md := metadata.Pairs("sign", sign, "pub", pub)
		if e := grpc.SetTrailer(ctx, md); e != nil {
			panic(err)
		}

		return resp, err
	}
}

// parseContext reads fields from request/response context.
func parseContext(ctx context.Context) (sign string, pub *common.PublicKey, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = errors.New("data should be signed")
		return
	}

	sign, pub, err = readMetadata(md)
	if err != nil {
		return
	}

	return
}

// readMetadata reads fields from metadata.
func readMetadata(md metadata.MD) (sign string, pub *common.PublicKey, err error) {
	signs, ok := md["sign"]
	if !ok || len(signs) < 1 {
		err = errors.New("data should be signed: no sign")
		return
	}
	sign = signs[0]

	pubs, ok := md["pub"]
	if !ok || len(pubs) < 1 {
		err = errors.New("data should be signed: no pub")
		return
	}
	pub, err = common.Base64ToPubkey(pubs[0])

	return
}

func signData(data interface{}, key *common.PrivateKey) string {
	h := hashOfData(data)

	R, S, _ := key.Sign(h.Bytes())

	return crypto.EncodeSignature(R, S)
}

func verifyData(data interface{}, sign string, pub *common.PublicKey) error {
	h := hashOfData(data)

	r, s, err := crypto.DecodeSignature(sign)
	if err != nil {
		return err
	}

	if !pub.Verify(h.Bytes(), r, s) {
		return errors.New("invalid signature")
	}

	return nil
}

func hashOfData(data interface{}) hash.Hash {
	d, ok := data.(proto.Message)
	if !ok {
		panic("data is not proto.Message")
	}

	if IsProtoEmpty(&d) {
		return hash.Hash{}
	}

	var pbf proto.Buffer
	pbf.SetDeterministic(true)
	if err := pbf.Marshal(d); err != nil {
		panic(err)
	}

	return hash.Of(pbf.Bytes())
}

// IsProtoEmpty return true if it is typed nil (by protobuf sources).
func IsProtoEmpty(m *proto.Message) bool {
	// Super-tricky - read pointer out of data word of interface value.
	// Saves ~25ns over the equivalent:
	// return valToPointer(reflect.ValueOf(*m))
	return m == nil || (*[2]unsafe.Pointer)(unsafe.Pointer(m))[1] == nil
}
