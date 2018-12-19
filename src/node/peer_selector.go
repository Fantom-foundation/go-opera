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

//+++++++++++++++++++++++++++++++++++++++
//RANDOM

// RandomPeerSelector is a randomized peer selection struct
type RandomPeerSelector struct {
	peers     *peers.Peers
	localAddr string
	last      string
}

// NewRandomPeerSelector creates a new random peer selector
func NewRandomPeerSelector(participants *peers.Peers, localAddr string) *RandomPeerSelector {
	return &RandomPeerSelector{
		localAddr: localAddr,
		peers:     participants,
	}
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
	selectablePeers := ps.peers.ToPeerSlice()

	if len(selectablePeers) > 1 {
		_, selectablePeers = peers.ExcludePeer(selectablePeers, ps.localAddr)

		if len(selectablePeers) > 1 {
			_, selectablePeers = peers.ExcludePeer(selectablePeers, ps.last)
		}
	}

	i := rand.Intn(len(selectablePeers))

	peer := selectablePeers[i]

	return peer
}
