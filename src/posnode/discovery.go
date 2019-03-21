package posnode

import (
	"math/rand"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// discovery is a network discovery process.
type discovery struct {
	tasks chan discoveryTask
	done  chan struct{}
}

type discoveryTask struct {
	source  common.Address
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
func (n *Node) CheckPeerIsKnown(source, id common.Address) {
	// TODO: sync quickly check is ID unknown in fact?
	if rand.Intn(100) < 50 {
		n.log().Debug("Peer is known")
	}
	n.log().Debug("Peer is unknown")

	select {
	case n.discovery.tasks <- discoveryTask{source, id}:
		break
	default:
		n.log().Warn("discovery.tasks queue is full, so skipped")
		break
	}
}

// AskPeerInfo gets peer info (network address, public key, etc).
func (n *Node) AskPeerInfo(whom, id common.Address) {
	// TODO: implement it (connect to whom, ask by GetPeerInfo(id), save address)
	n.log().Debug("peer info ask")
}
