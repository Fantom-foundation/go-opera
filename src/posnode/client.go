package posnode

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

const (
	// clientTimeout defines how long will gRPC client will wait
	// for response from the server.
	clientTimeout = 15 * time.Second
	// connectTimeout defines how long dialer will for connection
	// to be established.
	connectTimeout = 15 * time.Second
)

// client of node service.
// TODO: make reusable connections pool
type client struct {
	opts []grpc.DialOption
}

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(ctx context.Context, host string) (api.NodeClient, error) {
	var (
		conn *grpc.ClientConn
		err  error
	)

	addr := n.NetAddrOf(host)
	n.log.Debugf("connect to %s", addr)
	// TODO: secure connection
	conn, err = grpc.DialContext(ctx, addr, append(n.client.opts, grpc.WithInsecure())...)
	if err != nil {
		n.log.Error(errors.Wrapf(err, "connect to: %s", addr))
		return nil, err
	}

	return api.NewNodeClient(conn), nil
}
