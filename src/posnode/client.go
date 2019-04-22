package posnode

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// TODO: find better solution
// Temp data for client/server Interceptors
var (
	clientID  = ""
	clientKey = new(common.PrivateKey)
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
	clientKey = n.key

	addr := n.NetAddrOf(peer.Host)
	n.log.Debugf("connect to %s", addr)

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
	sign, pubKey := api.SignGRPCData(req, clientKey)

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
	id, sign, pub, err := api.GetInfoFromMetadata(trailer)
	if err != nil {
		return err
	}

	// Extract common.Pubkey from string
	key, err := common.StringToPubkey(pub)
	if err != nil {
		return err
	}

	// Validation ID
	if ok := api.ValidateID(id, key); !ok {
		return errors.New("ID is invalid")
	}

	// Check signature of response
	isValid, err := api.CheckSignData(req, sign, key)
	if err != nil {
		return err
	}

	if !isValid {
		return errors.New("Signature is invalid")
	}

	return nil
}
