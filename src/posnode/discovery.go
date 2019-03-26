package posnode

import (
	"context"
	"strconv"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
	"github.com/pkg/errors"
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
	source, unknown common.Address
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
func (n *Node) CheckPeerIsKnown(source, id common.Address, host string) {
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
func (n *Node) AskPeerInfo(source, id common.Address, host string) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	netAddr := host + ":" + strconv.Itoa(n.conf.Port)
	cli, err := n.ConnectTo(ctx, netAddr)
	if err != nil {
		n.log.Error(errors.Wrapf(err, "connect to: %s", netAddr))
		return
	}

	peerInfo, err := requestPeerInfo(cli, id.Hex())
	if err != nil {
		n.log.Error(errors.Wrapf(err, "request peer info: %s", netAddr))
		// TODO: handle not found.
		return
	}

	peer := WireToPeer(peerInfo)
	n.store.SetPeer(peer)
}

// requestPeerInfo makes GetPeerInfo using NodeClient
// with context which hash timeout.
func requestPeerInfo(cli wire.NodeClient, id string) (*wire.PeerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	in := wire.PeerRequest{
		PeerID: id,
	}
	return cli.GetPeerInfo(ctx, &in)
}
