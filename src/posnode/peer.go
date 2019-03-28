package posnode

import (
	"crypto/ecdsa"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

// Peer is a representation of other node.
type Peer struct {
	ID      common.Address
	PubKey  *ecdsa.PublicKey
	NetAddr string
}

// ToWire converts to protobuf message.
func (p *Peer) ToWire() *wire.PeerInfo {
	return &wire.PeerInfo{
		ID:      p.ID.Hex(),
		PubKey:  crypto.FromECDSAPub(p.PubKey),
		NetAddr: p.NetAddr,
	}
}

// WireToPeer converts from protobuf message.
func WireToPeer(w *wire.PeerInfo) *Peer {
	return &Peer{
		ID:      common.BytesToAddress(common.FromHex(w.ID)),
		PubKey:  crypto.ToECDSAPub(w.PubKey),
		NetAddr: w.NetAddr,
	}
}

func IDsToWire(ids []common.Address) *wire.PeersID {
	w := &wire.PeersID{
		IDs: make([]string, len(ids)),
	}

	for i, id := range ids {
		w.IDs[i] = id.Hex()
	}

	return w
}

func WireToIDs(w *wire.PeersID) []common.Address {
	if w == nil {
		return nil
	}

	res := make([]common.Address, len(w.IDs))
	for i, str := range w.IDs {
		res[i] = common.HexToAddress(str)
	}

	return res
}

// Connected is a representation of node address collection.
type Connected struct {
	mx sync.RWMutex
	m  map[common.Address]bool
}

// NewConnected create new Connected struct
func NewConnected() *Connected {
	return &Connected{
		m: make(map[common.Address]bool),
	}
}

// Load value about connected status by address
func (c *Connected) Load(key common.Address) bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	val, _ := c.m[key]

	return val
}

// Store value about connected status by address
func (c *Connected) Store(key common.Address, value bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.m[key] = value
}
