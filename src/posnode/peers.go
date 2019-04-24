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
	peers     map[hash.Peer]*peerAttr
	hosts     map[string]*hostAttr

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

func (pp *peers) attrByID(id hash.Peer) *peerAttr {
	if id.IsEmpty() {
		return nil
	}

	attr := pp.peers[id]
	if attr == nil {
		attr = &peerAttr{
			ID:   id,
			Host: &hostAttr{},
		}
		pp.peers[id] = attr
	}
	return attr
}

func (pp *peers) attrByHost(host string) *hostAttr {
	if host == "" {
		return nil
	}

	attr := pp.hosts[host]
	if attr == nil {
		attr = &hostAttr{
			Name: host,
		}
		pp.hosts[host] = attr
	}
	return attr
}

func (n *Node) initPeers() {
	n.initDownloads()
	n.initClient()

	if n.peers.top != nil {
		return
	}
	n.peers.top = n.store.GetTopPeers()
	n.peers.save = func() {
		n.store.SetTopPeers(n.peers.top)
	}

	n.peers.peers = make(map[hash.Peer]*peerAttr)
	n.peers.hosts = make(map[string]*hostAttr)
}

// ConnectOK counts successful connections to peer.
func (n *Node) ConnectOK(p *Peer) {
	n.peers.Lock()
	defer n.peers.Unlock()

	host := n.peers.attrByHost(p.Host)
	host.LastSuccess = time.Now()

	peer := n.peers.attrByID(p.ID)
	if peer == nil {
		return
	}
	peer.Host = host

	stored := n.store.GetPeer(p.ID)
	if stored == nil {
		return
	}
	if stored.Host != p.Host {
		stored.Host = p.Host
		n.store.SetPeer(stored)
	}

	for _, exist := range n.peers.top {
		if p.ID == exist {
			return
		}
	}

	n.peers.top = append(n.peers.top, p.ID)
	n.peers.unordered = true
	n.peers.save()
}

// ConnectFail counts unsuccessful connections to peer.
func (n *Node) ConnectFail(p *Peer, err error) {
	n.log.Warn(err)

	n.peers.Lock()
	defer n.peers.Unlock()

	host := n.peers.attrByHost(p.Host)
	host.LastFail = time.Now()

	peer := n.peers.attrByID(p.ID)
	if peer == nil {
		return
	}
	peer.Host = host

	n.peers.unordered = true
}

// PeerReadyForReq returns false if peer is not ready for request.
// TODO: test it
func (n *Node) PeerReadyForReq(host string) bool {
	n.peers.RLock()
	defer n.peers.RUnlock()

	attr := n.peers.attrByHost(host)

	return attr != nil &&
		(attr.LastFail.Before(attr.LastSuccess) ||
			attr.LastFail.Before(time.Now().Add(-n.conf.DiscoveryTimeout)))
}

// PeerUnknown returns true if peer should be discovered.
func (n *Node) PeerUnknown(id *hash.Peer) bool {
	if id.IsEmpty() {
		return true
	}

	n.peers.RLock()
	defer n.peers.RUnlock()

	attr := n.peers.attrByID(*id).Host

	return attr == nil ||
		(attr.LastSuccess.Before(time.Now().Add(-n.conf.DiscoveryTimeout)) &&
			attr.LastFail.Before(time.Now().Add(-n.conf.DiscoveryTimeout)))
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
				delete(n.peers.peers, id)
			}
			n.peers.top = n.peers.top[:peersCount]
			n.peers.save()
		}
	}

	// return first no busy
	for _, candidate := range n.peers.top {
		attrs := n.peers.attrByID(candidate)
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
	if p == nil {
		return
	}

	n.peers.Lock()
	defer n.peers.Unlock()

	n.peers.attrByID(p.ID).Busy = false
}
