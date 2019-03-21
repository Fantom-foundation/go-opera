package posnode

import (
	"context"
	"math"
	"net"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

type (
	// Dialer is a func for connecting to service.
	Dialer func(context.Context, string) (net.Conn, error)
)

// StartService starts node service.
func (n *Node) StartService(listener net.Listener) {
	n.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	wire.RegisterNodeServer(n.server, n)

	go func() {
		if err := n.server.Serve(listener); err != nil {
			// TODO: log error
		}
	}()
}

// StopService stops node service.
func (n *Node) StopService() {
	n.server.GracefulStop()
}

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(ctx context.Context, addr string) (wire.NodeClient, error) {
	var (
		conn *grpc.ClientConn
		err  error
	)
	if n.dialer == nil {
		conn, err = grpc.DialContext(ctx, addr)
	} else {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithContextDialer(n.dialer))
	}
	if err != nil {
		return nil, err
	}

	return wire.NewNodeClient(conn), nil
}

/*
 * wire.NodeServer implementation:
 */

// SyncEvents it remember their known events for future request
// and returns unknown for they events.
func (n *Node) SyncEvents(ctx context.Context, req *wire.KnownEvents) (*wire.KnownEvents, error) {
	return nil, nil
}

// GetEvent returns requested event.
func (n *Node) GetEvent(ctx context.Context, req *wire.EventRequest) (*wire.Event, error) {
	return nil, nil
}

// GetPeerInfo returns requested peer info.
func (n *Node) GetPeerInfo(ctx context.Context, req *wire.PeerRequest) (*wire.PeerInfo, error) {
	return nil, nil
}
