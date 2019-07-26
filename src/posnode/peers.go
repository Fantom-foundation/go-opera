package posnode

import (
	"sort"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// peers manages node peer list.
type peers struct {
	top       []hash.Peer
	unordered bool
	ids       map[hash.Peer]*peerAttr
	hosts     map[string]*hostAttr

	sync.RWMutex
	save func()
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

	n.peers.ids = make(map[hash.Peer]*peerAttr, n.conf.TopPeersCount)
	n.peers.hosts = make(map[string]*hostAttr, n.conf.TopPeersCount*4)
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

	attr := pp.ids[id]
	if attr == nil {
		attr = &peerAttr{
			ID:   id,
			Host: &hostAttr{},
		}
		pp.ids[id] = attr
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

// trimHosts trims host attrs cache.
// Args for easy tests.
func (n *Node) trimHosts(fromLen, toLen int) {
	n.peers.Lock()
	defer n.peers.Unlock()

	if len(n.peers.hosts) < fromLen {
		return
	}

	ordered := make([]*hostAttr, 0, len(n.peers.hosts))
	for _, attr := range n.peers.hosts {
		ordered = append(ordered, attr)
	}
	sort.Sort(hostsByTime(ordered))

	tail2del := ordered[toLen:]
	for _, h := range tail2del {
		delete(n.peers.hosts, h.Name)
	}
}

// ConnectOK counts successful connections to peer.
func (n *Node) ConnectOK(p *Peer) {
	n.peers.Lock()
	defer n.peers.Unlock()

	if ok := n.updatePeerAttrs(p, true); !ok {
		return
	}

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

	countNodePeersTop.Inc(1)
}

// ConnectFail counts unsuccessful connections to peer.
func (n *Node) ConnectFail(p *Peer, err error) {
	n.Warn(err)

	n.peers.Lock()
	defer n.peers.Unlock()

	if ok := n.updatePeerAttrs(p, false); !ok {
		return
	}

	n.peers.unordered = true
}

// PeerReadyForReq returns false if peer is not ready for discovery request.
func (n *Node) PeerReadyForReq(host string) bool {
	n.peers.RLock()
	defer n.peers.RUnlock()

	attr := n.peers.attrByHost(host)

	timeout := time.Now().Add(-n.conf.DiscoveryTimeout)
	errtimeout := time.Now().Add(-2 * n.conf.DiscoveryTimeout)

	return attr == nil ||
		(!attr.LastFail.After(attr.LastSuccess) || attr.LastFail.Before(errtimeout)) && attr.LastSuccess.Before(timeout)
}

// PeerUnknown returns true if peer should be discovered.
func (n *Node) PeerUnknown(id *hash.Peer) bool {
	if id.IsEmpty() {
		return true
	}

	n.peers.RLock()
	defer n.peers.RUnlock()

	attr := n.peers.attrByID(*id).Host

	timeout := time.Now().Add(-n.conf.DiscoveryTimeout)

	return attr == nil ||
		(attr.LastSuccess.Before(timeout) && attr.LastFail.Before(timeout))
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
		if len(n.peers.top) > n.conf.TopPeersCount {
			tail := n.peers.top[n.conf.TopPeersCount:]
			for _, id := range tail {
				delete(n.peers.ids, id)
			}
			n.peers.top = n.peers.top[:n.conf.TopPeersCount]
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

func (n *Node) updatePeerAttrs(p *Peer, isSuccess bool) bool {
	if p == nil {
		return false
	}

	host := n.peers.attrByHost(p.Host)
	if host == nil {
		return false
	}

	if isSuccess {
		host.LastSuccess = time.Now()
	} else {
		host.LastFail = time.Now()
	}

	peer := n.peers.attrByID(p.ID)
	if peer == nil {
		return false
	}

	peer.Host = host

	return true
}

/*
 * hosts sorting:
 */

// hostsByTime is for sorting.
type hostsByTime []*hostAttr

// Len is the number of elements in the collection.
func (hh hostsByTime) Len() int {
	return len(hh)
}

// Swap swaps the elements with indexes i and j.
func (hh hostsByTime) Swap(i, j int) { hh[i], hh[j] = hh[j], hh[i] }

// Less reports whether the element with
// index i should sort before the element with index j.
func (hh hostsByTime) Less(i, j int) bool {
	return hh[i].LastSuccess.After(hh[j].LastSuccess) ||
		hh[i].LastFail.Before(hh[j].LastFail)
}
