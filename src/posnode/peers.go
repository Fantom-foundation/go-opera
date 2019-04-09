package posnode

import (
	"sort"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

const peersCount = 10

// peers manages node peer list.
type peers struct {
	top       []hash.Peer
	unordered bool
	attrs     map[hash.Peer]*peerAttrs

	sync.RWMutex
	save func()
}

func (pp *peers) Snapshot() []hash.Peer {
	pp.RLock()
	defer pp.RUnlock()

	res := make([]hash.Peer, len(pp.top))
	copy(res, pp.top)
	return res
}

func (pp *peers) attrOf(id hash.Peer) *peerAttrs {
	attrs := pp.attrs[id]
	if attrs == nil {
		attrs = &peerAttrs{}
		pp.attrs[id] = attrs
	}
	return attrs
}

func (n *Node) initPeers() {
	if n.peers.top != nil {
		return
	}
	n.peers.top = n.store.GetTopPeers()
	n.peers.attrs = make(map[hash.Peer]*peerAttrs)
	n.peers.save = func() {
		n.store.SetTopPeers(n.peers.top)
	}
}

// ConnectOK counts successful connections to peer.
func (n *Node) ConnectOK(peer *Peer) {
	n.peers.Lock()
	defer n.peers.Unlock()

	attr := n.peers.attrOf(peer.ID)
	attr.LastSuccess = time.Now()
	attr.LastHost = peer.Host

	if peer.PubKey == nil {
		return
	}

	old := n.store.GetPeer(peer.ID)
	if old != nil && old.Host == peer.Host {
		return
	}
	n.store.SetPeer(peer)

	for _, exist := range n.peers.top {
		if peer.ID == exist {
			return
		}
	}

	n.peers.top = append(n.peers.top, peer.ID)
	n.peers.unordered = true
	n.peers.save()
}

// ConnectFail counts unsuccessful connections to peer.
func (n *Node) ConnectFail(peer *Peer, err error) {
	n.log.Warn(err)

	n.peers.Lock()
	defer n.peers.Unlock()

	attr := n.peers.attrOf(peer.ID)
	attr.LastFail = time.Now()
	attr.LastHost = peer.Host

	n.peers.unordered = true
}

// PeerReadyForReq returns false if peer is not ready for request.
// TODO: test it
func (n *Node) PeerReadyForReq(id hash.Peer, host string) bool {
	n.peers.RLock()
	defer n.peers.RUnlock()

	attr := n.peers.attrOf(id)

	if attr.LastHost == host &&
		attr.LastFail.After(attr.LastSuccess) &&
		attr.LastFail.After(time.Now().Add(-discoveryTimeout)) {
		return false
	}

	return true
}

// PeerUnknown returns true if peer should be discovered.
// TODO: test it
func (n *Node) PeerUnknown(id *hash.Peer) bool {
	if id == nil {
		return true
	}

	n.peers.RLock()
	defer n.peers.RUnlock()

	attr := n.peers.attrOf(*id)
	if attr.LastSuccess.After(time.Now().Add(-discoveryTimeout)) {
		return false
	}

	peer := n.store.GetPeer(*id)
	if peer == nil || peer.Host == "" || peer.PubKey == nil {
		return true
	}

	return false
}

// NextForGossip returns the best candidate to gossip with and marks it as busy.
// You should call FreePeer() to mark candidate as not busy.
func (n *Node) NextForGossip() *Peer {
	n.peers.Lock()
	defer n.peers.Unlock()

	if len(n.peers.top) < 1 {
		return nil
	}

	// order and trunc the top
	if n.peers.unordered {
		sort.Sort((*gossipEvaluation)(n))
		n.peers.unordered = false
		if len(n.peers.top) > peersCount {
			tail := n.peers.top[peersCount:]
			for _, id := range tail {
				delete(n.peers.attrs, id)
			}
			n.peers.top = n.peers.top[:peersCount]
			n.peers.save()
		}
	}

	// return first no busy
	for _, candidate := range n.peers.top {
		attrs := n.peers.attrOf(candidate)
		if !attrs.Busy {
			attrs.Busy = true
			peer := n.store.GetPeer(candidate)
			return peer
		}
	}

	return nil
}

// FreePeer marks peer as not busy.
func (n *Node) FreePeer(p *Peer) {
	n.peers.Lock()
	defer n.peers.Unlock()

	n.peers.attrOf(p.ID).Busy = false
}
