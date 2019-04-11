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
	Engage(peer string)  // indicate we are in communication with that peer
	Dismiss(peer string) // indicate we are not in communication with that peer
}

// RandomPeerSelector is a randomized peer selection struct
type RandomPeerSelector struct {
	peers     *peers.Peers
	localAddr string
	last      string
	pals      map[string]bool
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
		pals:      make(map[string]bool),
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
	// We need an exclusive access to ps.last for writing;
	// let use peers' lock instead of adding additional lock.
	// ps.last is accessed for read under peers' lock
	ps.peers.Lock()
	defer ps.peers.Unlock()

	ps.last = peer
}

// Next returns the next randomly selected peer(s) to communicate with
func (ps *RandomPeerSelector) Next() *peers.Peer {
	ps.peers.Lock()
	defer ps.peers.Unlock()

	slice := ps.peers.ToPeerSlice()
	selectablePeers := peers.ExcludePeers(slice, ps.localAddr, ps.last)

	for k, _ := range ps.pals {
		selectablePeers = peers.ExcludePeers(selectablePeers, k, k)
	}

	if len(selectablePeers) < 1 {
		selectablePeers = slice
	}

	i := rand.Intn(len(selectablePeers))

	peer := selectablePeers[i]

	return peer
}

// Indicate we are in communication with a peer
// so it would be excluded from next peer selection
func (ps *RandomPeerSelector) Engage(peer string) {
	ps.peers.Lock()
	defer ps.peers.Unlock()
	ps.pals[peer] = true
}

// Indicate we are not in communication with a peer
// so it could be selected as a next peer
func (ps *RandomPeerSelector) Dismiss(peer string) {
	ps.peers.Lock()
	defer ps.peers.Unlock()
	if _, ok := ps.pals[peer]; ok {
		delete(ps.pals, peer)
	}
}
