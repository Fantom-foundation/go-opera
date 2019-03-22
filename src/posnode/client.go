package posnode

import (
	"context"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

// client of node service.
//TODO: make reusable connections pool
type client struct {
	opts []grpc.DialOption
}

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(ctx context.Context, addr string) (wire.NodeClient, error) {
	var (
		conn *grpc.ClientConn
		err  error
	)
	// TODO: secure connection
	conn, err = grpc.DialContext(ctx, addr, append(n.client.opts, grpc.WithInsecure())...)
	if err != nil {
		return nil, err
	}

	return wire.NewNodeClient(conn), nil
}
