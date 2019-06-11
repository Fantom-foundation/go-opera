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

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// peerID is a internal key for context.Value().
type peerID struct{}

// ClientAuth makes client-side interceptor for identification.
func ClientAuth(key *crypto.PrivateKey, genesis hash.Hash) grpc.UnaryClientInterceptor {
	pub := key.Public().Base64()
	salt := genesis.Bytes()

	return func(ctx context.Context, method string, req interface{}, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// request:

		servSign := signData(req, key, salt)
		md := metadata.Pairs("sign", servSign, "pub", pub)
		ctx = metadata.NewOutgoingContext(ctx, md)

		var answer metadata.MD
		opts = append(opts, grpc.Trailer(&answer))
		err := invoker(ctx, method, req, resp, cc, opts...)
		if err != nil {
			return err
		}

		// response:

		servSign, servPub, err := readMetadata(answer)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}

		err = verifyData(resp, servSign, servPub, salt)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}

		if set, ok := ctx.Value(peerID{}).(func(hash.Peer)); ok {
			serverID := hash.PeerOfPubkey(servPub)
			set(serverID)
		}

		return nil
	}
}

// ServerAuth makes server-side interceptor for identification.
func ServerAuth(key *crypto.PrivateKey, genesis hash.Hash) grpc.UnaryServerInterceptor {
	pub := base64.StdEncoding.EncodeToString(key.Public().Bytes())
	salt := genesis.Bytes()

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// request:

		clientSign, clientPub, err := parseContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		err = verifyData(req, clientSign, clientPub, salt)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		clientID := hash.PeerOfPubkey(clientPub)
		ctx = context.WithValue(ctx, peerID{}, clientID)

		// response:

		resp, err := handler(ctx, req)

		sign := signData(resp, key, salt)
		md := metadata.Pairs("sign", sign, "pub", pub)
		if err := grpc.SetTrailer(ctx, md); err != nil {
			logger.Get().Fatal(err)
		}

		return resp, err
	}
}

// parseContext reads fields from request/response context.
func parseContext(ctx context.Context) (sign string, pub *crypto.PublicKey, err error) {
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
func readMetadata(md metadata.MD) (sign string, pub *crypto.PublicKey, err error) {
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
	pub, err = crypto.Base64ToPubKey(pubs[0])

	return
}

func signData(data interface{}, key *crypto.PrivateKey, salt []byte) string {
	h := hashOfData(data)

	d := append(h.Bytes(), salt...)

	R, S, _ := key.Sign(d)

	return crypto.EncodeSignature(R, S)
}

func verifyData(data interface{}, sign string, pub *crypto.PublicKey, salt []byte) error {
	h := hashOfData(data)

	d := append(h.Bytes(), salt...)

	r, s, err := crypto.DecodeSignature(sign)
	if err != nil {
		return err
	}

	if !pub.Verify(d, r, s) {
		return errors.New("signature is invalid or peer uses another genesis")
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
		logger.Get().Fatal(err)
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
