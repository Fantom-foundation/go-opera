package posnode

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// discovery is a network discovery process.
type discovery struct {
	tasks chan discoveryTask
	done  chan struct{}
}

// discoveryTask contains data required
// for peer discovery, source is a node
// from which we receive information
// about unknown node, host is a host
// address of source node.
type discoveryTask struct {
	source, unknown hash.Peer
	host            string
}

// StartDiscovery starts single thread network discovery.
// It should be called once.
func (n *Node) StartDiscovery() {
	n.discovery.tasks = make(chan discoveryTask, 100) // NOTE: magic buffer size
	n.discovery.done = make(chan struct{})

	go func() {
		for {
			select {
			case task := <-n.discovery.tasks:
				n.AskPeerInfo(task.source, task.unknown, task.host)
			case <-n.discovery.done:
				return
			}
		}
	}()
}

// StopDiscovery stops network discovery.
// It should be called once.
func (n *Node) StopDiscovery() {
	close(n.discovery.done)
}

// CheckPeerIsKnown checks peer is known otherwise makes discovery task.
func (n *Node) CheckPeerIsKnown(source, id hash.Peer, host string) {
	// Find peer by its id in storage.
	peerInfo := n.store.GetPeerInfo(id)
	if peerInfo != nil {
		// If peer found in storage - skip.
		return
	}

	select {
	case n.discovery.tasks <- discoveryTask{
		source:  source,
		unknown: id,
		host:    host,
	}:
	default:
		n.log.Warn("discovery.tasks queue is full, so skipped")
	}
}

// AskPeerInfo gets peer info (network address, public key, etc).
func (n *Node) AskPeerInfo(source, id hash.Peer, host string) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	cli, err := n.ConnectTo(ctx, host)
	if err != nil {
		return
	}

	peerInfo, err := requestPeerInfo(cli, id.Hex())
	if err != nil {
		n.log.Error(errors.Wrapf(err, "request peer info"))
		// TODO: handle not found.
		return
	}

	// TODO: check is it real host allowed connection?

	n.SetPeerHost(id, peerInfo.Host)

	peer := WireToPeer(peerInfo)
	n.store.SetPeer(peer)
}

// requestPeerInfo makes GetPeerInfo using NodeClient
// with context which hash timeout.
func requestPeerInfo(cli api.NodeClient, id string) (*api.PeerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	in := api.PeerRequest{
		PeerID: id,
	}
	return cli.GetPeerInfo(ctx, &in)
}
