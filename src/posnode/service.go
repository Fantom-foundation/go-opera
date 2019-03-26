package posnode

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

type service struct {
	server *grpc.Server
}

// StartService starts node service.
// It should be called once.
func (n *Node) StartService() {
	bind := net.JoinHostPort(n.host, strconv.Itoa(n.conf.Port))
	listener := network.TcpListener(bind)
	n.startService(listener)
}

func (n *Node) startService(listener net.Listener) {
	n.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	wire.RegisterNodeServer(n.server, n)

	n.log.Infof("service start at %v", listener.Addr())
	go func() {
		if err := n.server.Serve(listener); err != nil {
			n.log.Infof("service stop (%v)", err)
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
	id := common.HexToAddress(req.PeerID)
	peerInfo := n.store.GetPeerInfo(id)
	// peerInfo == nil means that peer not found
	// by given id
	if peerInfo == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("peer not found: %s", req.PeerID))
	}

	return peerInfo, nil
}
