package hash

import (
	"math/big"
)

// Peer is a unique peer identifier.
// It is a hash of peer's PubKey.
type Peer Hash

// EmptyPeer is empty peer identifier.
var EmptyPeer = Peer{}

// Bytes returns value as byte slice.
func (p *Peer) Bytes() []byte {
	return (*Hash)(p).Bytes()
}

// Big converts a hash to a big integer.
func (p *Peer) Big() *big.Int {
	return (*Hash)(p).Big()
}

// BytesToPeer converts bytes to peer id.
// If b is larger than len(h), b will be cropped from the left.
func BytesToPeer(b []byte) Peer {
	return Peer(FromBytes(b))
}

// Hex converts a hash to a hex string.
func (p *Peer) Hex() string {
	return (*Hash)(p).Hex()
}

// HexToPeer sets byte representation of s to peer id.
// If b is larger than len(h), b will be cropped from the left.
func HexToPeer(s string) Peer {
	return Peer(HexToHash(s))
}

// String returns human readable string representation.
func (p *Peer) String() string {
	if name := GetNodeName(*p); len(name) > 0 {
		return name
	}
	return (*Hash)(p).ShortString()
}

// IsEmpty returns true if hash is empty.
func (p *Peer) IsEmpty() bool {
	return p == nil || *p == EmptyPeer
}

/*
 * Utils:
 */

// FakePeer generates random fake peer id for testing purpose.
func FakePeer(seed ...int64) Peer {
	return Peer(FakeHash(seed...))
}
