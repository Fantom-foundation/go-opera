package posnode

import (
	"sort"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

const peersCount = 10

// peers manages node peer list.
type peers struct {
	top       []hash.Peer
	busy      map[hash.Peer]struct{}
	unordered bool

	sync sync.Mutex
	save func()
}

func initPeers(s *Store) peers {
	pp := peers{
		top:  s.GetTopPeers(),
		busy: make(map[hash.Peer]struct{}),
	}

	pp.save = func() {
		pp.sync.Lock()
		defer pp.sync.Unlock()
		s.SetTopPeers(pp.top)
	}

	return pp
}

// NextForGossip returns the best candidate to gossip with and marks it as busy.
// You should call FreePeer() to mark candidate as not busy.
func (n *Node) NextForGossip() *Peer {
	n.peers.sync.Lock()
	defer n.peers.sync.Unlock()

	if len(n.peers.top) < 1 {
		return nil
	}

	// order and trunc the top
	if n.peers.unordered {
		sort.Sort((*gossipEvaluation)(n))
		n.peers.unordered = false
		if len(n.peers.top) > peersCount {
			n.peers.top = n.peers.top[:peersCount]
			n.peers.save()
		}
	}

	// return first no busy
	for _, candidate := range n.peers.top {
		if _, ok := n.peers.busy[candidate]; !ok {
			peer := n.store.GetPeer(candidate)
			n.peers.busy[peer.ID] = struct{}{}
			return peer
		}
	}

	return nil
}

// FreePeer marks peer as not busy.
func (n *Node) FreePeer(p *Peer) {
	n.peers.sync.Lock()
	defer n.peers.sync.Unlock()

	delete(n.peers.busy, p.ID)
}

// SetPeerHost saves peer's host for gossip purpose.
func (n *Node) SetPeerHost(id hash.Peer, host string) {
	n.peers.sync.Lock()
	defer n.peers.sync.Unlock()

	peer := n.store.GetPeer(id)
	if peer != nil && peer.Host == host {
		return
	}
	if peer == nil {
		peer = &Peer{
			ID: id,
		}
	}

	peer.Host = host

	// if already exists
	for _, exist := range n.peers.top {
		if id == exist {
			return
		}
	}

	n.peers.top = append(n.peers.top, id)
	n.peers.unordered = true
	n.peers.save()
}
