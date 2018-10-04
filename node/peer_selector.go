package node

import (
	"math/rand"

	"github.com/andrecronje/lachesis/src/peers"
)

type PeerSelector interface {
	Peers() []*peers.Peer
	UpdateLast(peer string)
	Next() peers.Peer
}

//+++++++++++++++++++++++++++++++++++++++
//RANDOM

type RandomPeerSelector struct {
	peers []*peers.Peer
	last  string
}

func NewRandomPeerSelector(participants []*peers.Peer, localAddr string) *RandomPeerSelector {
	_, _peers := peers.ExcludePeer(participants, localAddr)
	return &RandomPeerSelector{
		peers: _peers,
	}
}

func (ps *RandomPeerSelector) Peers() []*peers.Peer {
	return ps.peers
}

func (ps *RandomPeerSelector) UpdateLast(peer string) {
	ps.last = peer
}

func (ps *RandomPeerSelector) Next() *peers.Peer {
	selectablePeers := ps.peers
	if len(selectablePeers) > 1 {
		_, selectablePeers = peers.ExcludePeer(selectablePeers, ps.last)
	}
	i := rand.Intn(len(selectablePeers))
	peer := selectablePeers[i]
	return peer
}
