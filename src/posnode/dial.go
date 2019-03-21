package posnode

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

// Dialer is a func for connecting to service.
type Dialer func(context.Context, string) (net.Conn, error)

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(ctx context.Context, addr string) (wire.NodeClient, error) {
	var (
		conn *grpc.ClientConn
		err  error
	)
	if n.peerDialer == nil {
		conn, err = grpc.DialContext(ctx, addr)
	} else {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithContextDialer(n.peerDialer))
	}
	if err != nil {
		return nil, err
	}

	return wire.NewNodeClient(conn), nil
}
