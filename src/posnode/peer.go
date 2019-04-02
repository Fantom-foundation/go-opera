package posnode

import (
	"crypto/ecdsa"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// Peer is a representation of other node.
type Peer struct {
	ID     hash.Peer
	PubKey *ecdsa.PublicKey
	Host   string
}

// ToWire converts to protobuf message.
func (p *Peer) ToWire() *api.PeerInfo {
	return &api.PeerInfo{
		ID:     p.ID.Hex(),
		PubKey: common.FromECDSAPub(p.PubKey),
		Host:   p.Host,
	}
}

// WireToPeer converts from protobuf message.
func WireToPeer(w *api.PeerInfo) *Peer {
	return &Peer{
		ID:     hash.HexToPeer(w.ID),
		PubKey: common.ToECDSAPub(w.PubKey),
		Host:   w.Host,
	}
}

func IDsToWire(ids []hash.Peer) *api.PeersID {
	w := &api.PeersID{
		IDs: make([]string, len(ids)),
	}

	for i, id := range ids {
		w.IDs[i] = id.Hex()
	}

	return w
}

func WireToIDs(w *api.PeersID) []hash.Peer {
	if w == nil {
		return nil
	}

	res := make([]hash.Peer, len(w.IDs))
	for i, str := range w.IDs {
		res[i] = hash.HexToPeer(str)
	}

	return res
}

// Connected is a representation of node address collection.
type Connected struct {
	mx sync.RWMutex
	m  map[hash.Peer]bool
}

// NewConnected create new Connected struct
func NewConnected() *Connected {
	return &Connected{
		m: make(map[hash.Peer]bool),
	}
}

// Load value about connected status by address
func (c *Connected) Load(key hash.Peer) bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	val, _ := c.m[key]

	return val
}

// Store value about connected status by address
func (c *Connected) Store(key hash.Peer, value bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.m[key] = value
}
