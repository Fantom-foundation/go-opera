package node

import (
	"math"
	"math/rand"

	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

// FairPeerSelector provides selection to prevent lazy node creation
type FairPeerSelector struct {
	// kPeerSize uint64
	last      string
	localAddr string
	peers     *peers.Peers
	pals      map[string]bool
}

// FairPeerSelectorCreationFnArgs specifies which additional arguments are require to create a FairPeerSelector
type FairPeerSelectorCreationFnArgs struct {
	KPeerSize uint64
	LocalAddr string
}

// NewFairPeerSelector creates a new fair peer selection struct
func NewFairPeerSelector(participants *peers.Peers, args FairPeerSelectorCreationFnArgs) *FairPeerSelector {
	return &FairPeerSelector{
		localAddr: args.LocalAddr,
		peers:     participants,
		pals:      make(map[string]bool),
		// kPeerSize: args.KPeerSize,
	}
}

// NewFairPeerSelectorWrapper implements SelectorCreationFn to allow dynamic creation of FairPeerSelector ie NewNode
func NewFairPeerSelectorWrapper(participants *peers.Peers, args interface{}) PeerSelector {
	return NewFairPeerSelector(participants, args.(FairPeerSelectorCreationFnArgs))
}

// Peers returns all known peers
func (ps *FairPeerSelector) Peers() *peers.Peers {
	return ps.peers
}

// UpdateLast sets the last peer communicated with (avoid double talk)
func (ps *FairPeerSelector) UpdateLast(peer string) {
	// We need exclusive access to ps.last for writing;
	// let use peers' lock instead of adding an additional lock.
	// ps.last is accessed for read under peers' lock
	ps.peers.Lock()
	defer ps.peers.Unlock()

	ps.last = peer
}

func fairCostFunction(peer *peers.Peer) float64 {
	if peer.Height == 0 {
		return 0
	}
	return float64(peer.InDegree / peer.Height)
}

// Next returns the next peer based on the work cost function selection
func (ps *FairPeerSelector) Next() *peers.Peer {
	// Maximum number of peers to select/return. In case configurable KPeerSize is implemented.
	// maxPeers := ps.kPeerSize
	// if maxPeers == 0 {
	// 	maxPeers = 1
	// }

	ps.peers.Lock()
	defer ps.peers.Unlock()

	sortedSrc := ps.peers.ToPeerByUsedSlice()
	var lastUsed []*peers.Peer

	minCost := math.Inf(1)
	var selected []*peers.Peer
	for _, p := range sortedSrc {
		if p.NetAddr == ps.localAddr {
			continue
		}
		if p.NetAddr == ps.last || p.PubKeyHex == ps.last {
			lastUsed = append(lastUsed, p)
			continue
		}
		// skip peers we are alredy engaged with
		if _, ok := ps.pals[p.NetAddr]; ok {
			continue
		}

		cost := fairCostFunction(p)
		if minCost > cost {
			minCost = cost
			selected = make([]*peers.Peer, 1)
			selected[0] = p
		} else if minCost == cost {
			selected = append(selected, p)
		}

	}

	if len(selected) < 1 {
		selected = lastUsed
	}
	if len(selected) == 1 {
		selected[0].Used++
		return selected[0]
	}
	if len(selected) < 1 {
		return nil
	}

	i := rand.Intn(len(selected))
	selected[i].Used++
	return selected[i]
}

// Indicate we are in communication with a peer
// so it would be excluded from next peer selection
func (ps *FairPeerSelector) Engage(peer string) {
	ps.peers.Lock()
	defer ps.peers.Unlock()
	ps.pals[peer] = true
}

// Indicate we are not in communication with a peer
// so it could be selected as a next peer
func (ps *FairPeerSelector) Dismiss(peer string) {
	ps.peers.Lock()
	defer ps.peers.Unlock()
	if _, ok := ps.pals[peer]; ok {
		delete(ps.pals, peer)
	}
}
