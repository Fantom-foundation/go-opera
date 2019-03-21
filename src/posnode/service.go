package posnode

import (
	"context"
	"math"
	"net"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

type service struct {
	server *grpc.Server
}

// StartService starts node service.
// It should be called once.
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
// It should be called once.
func (n *Node) StopService() {
	n.server.GracefulStop()
}

/*
 * wire.NodeServer implementation:
 */

// SyncEvents it remember their known events for future request
// and returns unknown for they events.
func (n *Node) SyncEvents(ctx context.Context, req *wire.KnownEvents) (*wire.KnownEvents, error) {
	// TODO: implement it
	return nil, nil
}

// GetEvent returns requested event.
func (n *Node) GetEvent(ctx context.Context, req *wire.EventRequest) (*wire.Event, error) {
	// TODO: implement it
	return nil, nil
}

// GetPeerInfo returns requested peer info.
func (n *Node) GetPeerInfo(ctx context.Context, req *wire.PeerRequest) (*wire.PeerInfo, error) {
	// TODO: implement it
	return nil, nil
}
