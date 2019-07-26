package api

import (
	"context"
	"errors"
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// peerID is a internal key for context.Value().
type peerID struct{}

// ClientAuth makes client-side interceptor for identification.
func ClientAuth(key *crypto.PrivateKey, genesis hash.Hash) grpc.UnaryClientInterceptor {
	addr := cryptoaddr.AddressOf(key.Public())
	salt := genesis.Bytes()

	return func(ctx context.Context, method string, req interface{}, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// request:

		servSign := signData(req, key, salt)
		md := metadata.Pairs("sign", common.Bytes2Hex(servSign), "addr", addr.Hex())
		ctx = metadata.NewOutgoingContext(ctx, md)

		var answer metadata.MD
		opts = append(opts, grpc.Trailer(&answer))
		err := invoker(ctx, method, req, resp, cc, opts...)
		if err != nil {
			return err
		}

		// response:

		servSig, servAddr, err := readMetadata(answer)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}

		err = verifyData(resp, servSig, servAddr, salt)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}

		if set, ok := ctx.Value(peerID{}).(func(hash.Peer)); ok {
			set(servAddr)
		}

		return nil
	}
}

// ServerAuth makes server-side interceptor for identification.
func ServerAuth(key *crypto.PrivateKey, genesis hash.Hash) grpc.UnaryServerInterceptor {
	addr := cryptoaddr.AddressOf(key.Public())
	salt := genesis.Bytes()

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// request:

		clientSign, clientAddr, err := parseContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		err = verifyData(req, clientSign, clientAddr, salt)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		ctx = context.WithValue(ctx, peerID{}, clientAddr)

		// response:

		resp, err := handler(ctx, req)

		sig := signData(resp, key, salt)
		md := metadata.Pairs("sign", common.Bytes2Hex(sig), "addr", addr.Hex())
		if err := grpc.SetTrailer(ctx, md); err != nil {
			logger.Get().Fatal(err)
		}

		return resp, err
	}
}

// parseContext reads fields from request/response context.
func parseContext(ctx context.Context) (sig []byte, addr hash.Peer, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = errors.New("data should be signed")
		return
	}

	sig, addr, err = readMetadata(md)
	if err != nil {
		return
	}

	return
}

// readMetadata reads fields from metadata.
func readMetadata(md metadata.MD) (sig []byte, addr hash.Peer, err error) {
	signs, ok := md["sign"]
	if !ok || len(signs) < 1 {
		err = errors.New("data should be signed: no sign")
		return
	}
	sig = common.Hex2Bytes(signs[0])

	addrs, ok := md["addr"]
	if !ok || len(addrs) < 1 {
		err = errors.New("data should be signed: no addr")
		return
	}
	addr = hash.HexToPeer(addrs[0])

	return
}

func signData(data interface{}, key *crypto.PrivateKey, salt []byte) []byte {
	h := hashOfData(data)

	salted := crypto.Keccak256(append(h.Bytes(), salt...))

	sig, _ := key.Sign(salted)

	return sig
}

func verifyData(data interface{}, sig []byte, addr hash.Peer, salt []byte) error {
	h := hashOfData(data)

	salted := hash.FromBytes(crypto.Keccak256(append(h.Bytes(), salt...)))

	if !cryptoaddr.VerifySignature(addr, salted, sig) {
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
