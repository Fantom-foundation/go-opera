package posnode

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// TODO: find better solution
// Temp data for client/server Interceptors
var (
	clientID = ""
	key      = new(common.PrivateKey)
)

type (
	// client of node service.
	// TODO: make reusable connections pool
	client struct {
		opts []grpc.DialOption
	}
)

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(peer *Peer) (api.NodeClient, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), n.conf.ConnectTimeout)
	defer cancel()

	// Fill temp data for Interceptor
	clientID = n.ID.Hex()
	key = n.key

	addr := n.NetAddrOf(peer.Host)
	n.log.Debugf("connect to %s", addr)
	// TODO: secure connection
	conn, err := grpc.DialContext(ctx, addr, append(n.client.opts, grpc.WithInsecure(), grpc.WithUnaryInterceptor(clientInterceptor), grpc.WithBlock())...)
	if err != nil {
		n.log.Warn(errors.Wrapf(err, "connect to: %s", addr))
		return nil, nil, err
	}

	free := func() {
		conn.Close()
	}

	return api.NewNodeClient(conn), free, nil
}

func clientInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Sign request
	sign, pubKey := signRequest(req)

	// Create new metadata for current request
	md := metadata.Pairs("client_id", clientID, "client_sign", sign, "client_pub", pubKey)

	// Append new metadata to context
	ctx = metadata.NewOutgoingContext(ctx, md)

	return invoker(ctx, method, req, reply, cc, opts...)
}

func signRequest(req interface{}) (sign, pubKey string) {
	b, _ := req.([]byte)
	R, S, _ := key.Sign(b)

	sign = crypto.EncodeSignature(R, S)
	pubKey = fmt.Sprint(key.Public().Bytes())

	return
}
