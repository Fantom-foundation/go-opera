package node

import (
	"math/rand"

	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

// PeerSelector provides an interface for the lachesis node to
// update the last peer it gossiped with and select the next peer
// to gossip with
type PeerSelector interface {
	Peers() *peers.Peers
	UpdateLast(peer string)
	Next() *peers.Peer
}

// RandomPeerSelector is a randomized peer selection struct
type RandomPeerSelector struct {
	peers     *peers.Peers
	localAddr string
	last      string
}

// SelectorCreationFnArgs specifies the union of possible arguments that can be extracted to create a variant of PeerSelector
type SelectorCreationFnArgs interface{}

// SelectorCreationFn declares the function signature to create variants of PeerSelector
type SelectorCreationFn func(*peers.Peers, interface{}) PeerSelector

// RandomPeerSelectorCreationFnArgs arguments for RandomPeerSelector
type RandomPeerSelectorCreationFnArgs struct {
	LocalAddr string
}

// NewRandomPeerSelector creates a new random peer selector
func NewRandomPeerSelector(participants *peers.Peers, args RandomPeerSelectorCreationFnArgs) *RandomPeerSelector {
	return &RandomPeerSelector{
		localAddr: args.LocalAddr,
		peers:     participants,
	}
}

// NewRandomPeerSelectorWrapper implements SelectorCreationFn to allow dynamic creation of RandomPeerSelector ie NewNode
func NewRandomPeerSelectorWrapper(participants *peers.Peers, args interface{}) PeerSelector {
	return NewRandomPeerSelector(participants, args.(RandomPeerSelectorCreationFnArgs))
}

// Peers returns all known peers
func (ps *RandomPeerSelector) Peers() *peers.Peers {
	return ps.peers
}

// UpdateLast sets the last peer communicated with (to avoid double talk)
func (ps *RandomPeerSelector) UpdateLast(peer string) {
	ps.last = peer
}

// Next returns the next randomly selected peer(s) to communicate with
func (ps *RandomPeerSelector) Next() *peers.Peer {
	slice := ps.peers.ToPeerSlice()
	selectablePeers := peers.ExcludePeers(slice, ps.localAddr, ps.last)

	if len(selectablePeers) < 1 {
		selectablePeers = slice
	}

	i := rand.Intn(len(selectablePeers))

	peer := selectablePeers[i]

	return peer
}
