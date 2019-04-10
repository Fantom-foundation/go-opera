package posnode

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// client of node service.
// TODO: make reusable connections pool
type client struct {
	opts []grpc.DialOption
}

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(peer *Peer) (api.NodeClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), n.conf.ConnectTimeout)

	addr := n.NetAddrOf(peer.Host)
	n.log.Debugf("connect to %s", addr)
	// TODO: secure connection
	conn, err := grpc.DialContext(ctx, addr, append(n.client.opts, grpc.WithInsecure())...)
	if err != nil {
		n.log.Warn(errors.Wrapf(err, "connect to: %s", addr))
		return nil, err
	}

	return api.NewNodeClient(conn), nil
}
