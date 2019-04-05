package posnode

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

const (
	// waitOnDiscoveryFailure is the how often discovery should try to request.
	discoveryTimeout = 5 * time.Minute
)

type (
	// discovery is a network discovery process.
	discovery struct {
		tasks chan discoveryTask
		done  chan struct{}
	}

	// discoveryTask is a task to ask source by host for unknown peer.
	discoveryTask struct {
		source  hash.Peer
		host    string
		unknown *hash.Peer
	}
)

// StartDiscovery starts single thread network discovery.
func (n *Node) StartDiscovery() {
	if n.discovery.done != nil {
		return
	}

	n.initPeers()

	n.discovery.tasks = make(chan discoveryTask, 100) // magic buffer size.
	n.discovery.done = make(chan struct{})

	go func() {
		for {
			select {
			case task := <-n.discovery.tasks:
				n.AskPeerInfo(task.source, task.host, task.unknown)
			case <-n.discovery.done:
				return
			}
		}
	}()
}

// StopDiscovery stops network discovery.
func (n *Node) StopDiscovery() {
	close(n.discovery.done)
	n.discovery.done = nil
}

// CheckPeerIsKnown queues peer checking for a late.
func (n *Node) CheckPeerIsKnown(source hash.Peer, host string, id hash.Peer) {
	select {
	case n.discovery.tasks <- discoveryTask{
		source:  source,
		host:    host,
		unknown: &id,
	}:
	default:
		n.log.Warn("discovery.tasks queue is full, so skipped")
	}
}

// AskPeerInfo gets peer info (network address, public key, etc).
func (n *Node) AskPeerInfo(source hash.Peer, host string, id *hash.Peer) {
	if !n.PeerReadyForReq(source, host) {
		return
	}
	if !n.PeerUnknown(id) {
		return
	}

	peer := n.store.GetPeer(source)
	if peer == nil {
		peer = &Peer{ID: source}
	}
	peer.Host = host

	client, err := n.ConnectTo(peer)
	if err != nil {
		n.ConnectFail(peer, err)
		return
	}

	info, err := n.requestPeerInfo(client, id)
	if err != nil {
		n.ConnectFail(peer, err)
		return
	}

	if id == nil {
		n.ConnectOK(peer)
		return
	}

	if info == nil {
		n.log.Warnf("peer %s (%s) knows nothing about %s", source.String(), host, id.String())
		n.ConnectOK(peer)
		return
	}

	got := CalcPeerInfoID(info.PubKey)
	if got != *id {
		n.ConnectFail(peer, fmt.Errorf("bad PeerInfo response"))
		return
	}
	n.ConnectOK(peer)

	info.ID = got.Hex()
	n.store.SetWirePeer(*id, info)

	n.AskPeerInfo(*id, info.Host, nil)
}

// requestPeerInfo does GetPeerInfo request.
func (n *Node) requestPeerInfo(client api.NodeClient, id *hash.Peer) (info *api.PeerInfo, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req := api.PeerRequest{}
	if id != nil {
		req.PeerID = id.Hex()
	}

	info, err = client.GetPeerInfo(ctx, &req)
	if err == nil {
		return
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		info, err = nil, nil
	}
	return
}
