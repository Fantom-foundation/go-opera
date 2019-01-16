package peers

import (
	"encoding/hex"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

const (
	jsonPeerPath = "peers.json"
)

// NewPeer creates a new peer based on public key and network address
func NewPeer(pubKeyHex, netAddr string) *Peer {
	peer := &Peer{
		PubKeyHex: pubKeyHex,
		NetAddr:   netAddr,
		Used:      0,
	}

	peer.computeID()

	return peer
}

// Equals checks peers for equality
func (p *Peer) Equals(cmp *Peer) bool {
	return p.ID == cmp.ID &&
		p.NetAddr == cmp.NetAddr &&
		p.PubKeyHex == cmp.PubKeyHex
}

// PubKeyBytes returns the public key bytes for a peer
func (p *Peer) PubKeyBytes() ([]byte, error) {
	return hex.DecodeString(p.PubKeyHex[2:])
}

func (p *Peer) computeID() error {
	// TODO: Use the decoded bytes from hex
	pubKey, err := p.PubKeyBytes()

	if err != nil {
		return err
	}

	p.ID = int64(common.Hash32(pubKey))

	return nil
}

// Address returns the address for a peer
// TODO: hash of publickey
func (p *Peer) Address() (a common.Address) {
	bytes, err := p.PubKeyBytes()
	if err != nil {
		panic(err)
	}
	copy(a[:], bytes)
	return
}

// PeerStore provides an interface for persistent storage and
// retrieval of peers.
type PeerStore interface {
	// Peers returns the list of known peers.
	Peers() (*Peers, error)

	// SetPeers sets the list of known peers. This is invoked when a peer is
	// added or removed.
	SetPeers([]*Peer) error
}

// ExcludePeer is used to exclude a single peer from a list of peers.
func ExcludePeer(peers []*Peer, peer string) (int, []*Peer) {
	index := -1
	otherPeers := make([]*Peer, 0, len(peers))
	for i, p := range peers {
		if p.NetAddr != peer && p.PubKeyHex != peer {
			otherPeers = append(otherPeers, p)
		} else {
			index = i
		}
	}
	return index, otherPeers
}

// ExcludePeers is used to exclude multiple peers from a list of peers.
func ExcludePeers(peers []*Peer, local string, last string) []*Peer {
	otherPeers := make([]*Peer, 0, len(peers))
	for _, p := range peers {
		if p.NetAddr != local &&
			p.PubKeyHex != local &&
			p.NetAddr != last &&
			p.PubKeyHex != last {
			otherPeers = append(otherPeers, p)
		}
	}
	return otherPeers
}
