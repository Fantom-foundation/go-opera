package api

import (
	"context"
	"encoding/base64"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// ClientInterceptor interceptor for client
func ClientInterceptor(clientID string, clientKey *common.PrivateKey) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Sign request
		sign, pubKey := signGRPCData(req, clientKey)

		// Create new metadata for current request
		md := metadata.Pairs("id", clientID, "sign", sign, "pub", pubKey)

		// Append new metadata to context
		ctx = metadata.NewOutgoingContext(ctx, md)

		var trailer metadata.MD
		opts = append(opts, grpc.Trailer(&trailer))

		// Process request
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			return err
		}

		// Get metadata from response
		id, sign, pub, err := getInfoFromMetadata(trailer)
		if err != nil {
			return err
		}

		// Validate data
		err = validateData(reply, id, sign, pub)
		if err != nil {
			return err
		}

		return nil
	}
}

// serverInterceptor interceptor for server
func serverInterceptor(serverID string, serverKey *common.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get metadata from context
		id, sign, pub, err := getMetadata(ctx)
		if err != nil {
			return nil, err
		}

		// Validate data
		err = validateData(req, id, sign, pub)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}

		// Process request
		resp, err := handler(ctx, req)

		// Sign response
		sign, pubKey := signGRPCData(resp, serverKey)

		// Create new metadata for current response
		md := metadata.Pairs("id", serverID, "sign", sign, "pub", pubKey)
		grpc.SetTrailer(ctx, md)

		return resp, err
	}
}

func validateData(data interface{}, id, sign, pub string) error {
	// Extract common.Pubkey from string
	key, err := common.StringToPubkey(pub)
	if err != nil {
		return err
	}

	// Validation ID
	if ok := validateID(id, key); !ok {
		return errors.New("ID is invalid")
	}

	// Check signature of request / response
	isValid, err := checkSignData(data, sign, key)
	if err != nil {
		return err
	}

	if !isValid {
		return errors.New("Signature is invalid")
	}

	return nil
}

// getMetadata from context
func getMetadata(ctx context.Context) (rID, rSign, rPub string, err error) {
	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
		return
	}

	rID, rSign, rPub, err = getInfoFromMetadata(md)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, err.Error())
		return
	}

	return
}

// getInfoFromMetadata get info from request/response metadata
func getInfoFromMetadata(md metadata.MD) (rID, rSign, rPub string, err error) {
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

// validateID from response / request
func validateID(id string, key *common.PublicKey) bool {
	p := hash.PeerOfPubkey(key)
	return p.Hex() == id
}

// signGRPCData before send it
func signGRPCData(data interface{}, key *common.PrivateKey) (sign, pubKey string) {
	b, _ := data.([]byte)
	R, S, _ := key.Sign(b)

	sign = crypto.EncodeSignature(R, S)
	pubKey = base64.StdEncoding.EncodeToString(key.Public().Bytes())

	return
}

// checkSignData from response / request
func checkSignData(data interface{}, sign string, key *common.PublicKey) (bool, error) {
	hash, _ := data.([]byte)
	r, s, err := crypto.DecodeSignature(sign)
	if err != nil {
		return false, err
	}

	return key.Verify(hash, r, s), nil
}
