package posnode

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

type service struct {
	Addr       string
	listen     network.ListenFunc
	stopServer func()
}

// StartService starts node service.
func (n *Node) StartService() {
	if n.stopServer != nil {
		return
	}

	n.initLasts()

	if n.service.listen == nil {
		n.service.listen = network.TCPListener
	}

	var genesis hash.Hash
	if n.consensus != nil {
		genesis = n.consensus.GetGenesisHash()
	}

	bind := n.NetAddrOf(n.host)
	_, n.Addr, n.stopServer = api.StartService(bind, n.key, genesis, n, n.Infof, n.service.listen)
}

// StopService stops node service.
func (n *Node) StopService() {
	if n.stopServer == nil {
		return
	}
	n.stopServer()
	n.stopServer = nil
}

/*
 * api.NodeServer implementation:
 */

// SyncEvents returns their known event heights excluding heights from request.
func (n *Node) SyncEvents(ctx context.Context, req *api.KnownEvents) (*api.KnownEvents, error) {
	if err := checkSource(ctx); err != nil {
		return nil, err
	}

	// food for discovery
	host := api.GrpcPeerHost(ctx)
	n.CheckPeerIsKnown(host, nil)
	for hex := range req.Lasts {
		peer := hash.HexToPeer(hex)
		n.CheckPeerIsKnown(host, &peer)
	}

	// response
	knowns := n.knownEvents(idx.SuperFrame(req.SuperFrameN))

	return &api.KnownEvents{
		SuperFrameN: req.SuperFrameN,
		Lasts:       knowns.ToWire(),
	}, nil
}

// GetEvent returns requested event.
func (n *Node) GetEvent(ctx context.Context, req *api.EventRequest) (*wire.Event, error) {
	if err := checkSource(ctx); err != nil {
		return nil, err
	}

	// food for discovery
	host := api.GrpcPeerHost(ctx)
	n.CheckPeerIsKnown(host, nil)

	var eventHash hash.Event

	if req.Hash != nil {
		eventHash.SetBytes(req.Hash)
	} else {
		creator := hash.HexToPeer(req.PeerID)
		h := n.store.GetEventHash(creator, idx.SuperFrame(req.SuperFrameN), idx.Event(req.Seq))
		if h == nil {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("event not found: %s-%d-%d", req.PeerID, req.SuperFrameN, req.Seq))
		}
		eventHash = *h
	}

	event := n.store.GetWireEvent(eventHash)
	if event == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("event not found: %s", eventHash.Hex()))
	}

	return event, nil
}

// GetPeerInfo returns requested peer info.
func (n *Node) GetPeerInfo(ctx context.Context, req *api.PeerRequest) (*api.PeerInfo, error) {
	if err := checkSource(ctx); err != nil {
		return nil, err
	}

	// food for discovery
	host := api.GrpcPeerHost(ctx)
	n.CheckPeerIsKnown(host, nil)

	var id hash.Peer

	if req.PeerID == "" {
		id = n.ID
	} else {
		id = hash.HexToPeer(req.PeerID)
	}

	if id == n.ID { // self
		info := n.AsPeer()
		return info.ToWire(), nil
	}

	info := n.store.GetWirePeer(id)
	if info == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("peer not found: %s", req.PeerID))
	}

	return info, nil
}

/*
 * Utils:
 */

func checkSource(ctx context.Context) error {
	source := api.GrpcPeerID(ctx)
	if source.IsEmpty() {
		return status.Error(codes.Unauthenticated, "unknown peer")
	}
	return nil
}
