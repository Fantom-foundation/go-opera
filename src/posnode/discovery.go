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
	discoveryIdle = time.Second * 5
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
	done := n.discovery.done

	go func() {
		for {
			select {
			case task := <-n.discovery.tasks:
				n.AskPeerInfo(task.source, task.host, task.unknown)
			case <-time.After(discoveryIdle):
				if host := n.NextBuiltInPeer(); host != "" {
					n.AskPeerInfo(hash.EmptyPeer, host, nil)
				}
			case <-done:
				return
			}
		}
	}()

	n.log.Info("discovery started")
}

// StopDiscovery stops network discovery.
func (n *Node) StopDiscovery() {
	if n.discovery.done == nil {
		return
	}

	close(n.discovery.done)
	n.discovery.done = nil

	n.log.Info("discovery stopped")
}

// CheckPeerIsKnown queues peer checking for a late.
func (n *Node) CheckPeerIsKnown(source hash.Peer, host string, id *hash.Peer) {
	select {
	case n.discovery.tasks <- discoveryTask{
		source:  source,
		host:    host,
		unknown: id,
	}:
	default:
		n.log.Warn("discovery.tasks queue is full, so skipped")
	}
}

// AskPeerInfo gets peer info (network address, public key, etc).
func (n *Node) AskPeerInfo(source hash.Peer, host string, id *hash.Peer) {
	if !n.PeerReadyForReq(host) {
		return
	}
	if !n.PeerUnknown(id) {
		return
	}

	n.log.Debugf("ask %s about peer %s", host, id)

	peer := &Peer{Host: host}

	client, free, err := n.ConnectTo(peer)
	if err != nil {
		n.ConnectFail(peer, err)
		return
	}
	defer free()

	info, err := n.requestPeerInfo(client, id)
	if err != nil {
		n.ConnectFail(peer, err)
		return
	}

	if info == nil {
		if id == nil {
			n.ConnectFail(peer, fmt.Errorf("host %s knows nothing about self", host))
		} else {
			n.log.Warnf("peer %s (%s) knows nothing about %s", source.String(), host, id.String())
			n.ConnectOK(peer)
		}
		return
	}

	if hash.PeerOfPubkeyBytes(info.PubKey) != hash.HexToPeer(info.ID) {
		n.ConnectFail(peer, fmt.Errorf("bad PeerInfo response"))
		return
	}

	if id != nil && source != *id {
		n.ConnectOK(peer)
		n.AskPeerInfo(*id, info.Host, nil)
		return
	}

	info.Host = host
	peer = WireToPeer(info)
	n.store.SetWirePeer(peer.ID, info)
	n.ConnectOK(peer)
}

// requestPeerInfo does GetPeerInfo request.
func (n *Node) requestPeerInfo(client api.NodeClient, id *hash.Peer) (info *api.PeerInfo, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), n.conf.ClientTimeout)
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
