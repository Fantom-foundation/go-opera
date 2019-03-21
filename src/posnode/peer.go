package posnode

import (
	"crypto/ecdsa"

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
