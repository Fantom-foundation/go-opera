package posnode

import (
	"context"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
	"github.com/pkg/errors"
)

// discovery is a network discovery process.
type discovery struct {
	tasks chan discoveryTask
	done  chan struct{}
}

type discoveryTask struct {
	source  string // NetAddr
	unknown common.Address
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
				n.AskPeerInfo(task.source, task.unknown)
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
func (n *Node) CheckPeerIsKnown(source string, id common.Address) {
	// Find peer by its id in storage.
	_, err := n.store.GetPeerInfo(id)
	if err == nil {
		// If peer found in storage skip.
		return
	}

	if errors.Cause(err) != kvdb.ErrKeyNotFound {
		// Some unknown error with database, log and skip.
		// TODO: log error.
		return
	}

	select {
	case n.discovery.tasks <- discoveryTask{source, id}:
	default:
		n.log.Warn("discovery.tasks queue is full, so skipped")
	}
}

// AskPeerInfo gets peer info (network address, public key, etc).
func (n *Node) AskPeerInfo(whom string, id common.Address) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	cli, err := n.ConnectTo(ctx, whom)
	if err != nil {
		// TODO: redo task?
		// TODO: log error.
		return
	}

	peerInfo, err := requestPeerInfo(cli, id.Hex())
	if err != nil {
		// TODO: handle not found.
		// TODO: log error.
		return
	}

	peer := WireToPeer(peerInfo)
	if err := n.store.SetPeer(peer); err != nil {
		// TODO: log error.
		return
	}
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
