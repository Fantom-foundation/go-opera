package hash

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

var (
	// NodeNameDict is an optional dictionary to make node address human readable in log.
	NodeNameDict = make(map[Peer]string)
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

// PeerOfPubkeyBytes calcs peer id from pub key bytes.
func PeerOfPubkeyBytes(b []byte) Peer {
	return Peer(Of(b))
}

// PeerOfPubkey calcs peer id from pub key.
func PeerOfPubkey(pub *crypto.PublicKey) Peer {
	return Peer(Of(pub.Bytes()))
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
	if name, ok := NodeNameDict[*p]; ok {
		return name
	}
	return (*Hash)(p).ShortString()
}

// UnmarshalJSON parses a hash in hex syntax.
func (p *Peer) UnmarshalJSON(input []byte) error {
	return (*Hash)(p).UnmarshalJSON(input)
}

// IsEmpty returns true if hash is empty.
func (p *Peer) IsEmpty() bool {
	return p == nil || *p == EmptyPeer
}

/*
 * Utils:
 */

// FakePeer generates random fake peer id for testing purpose.
func FakePeer() Peer {
	return Peer(FakeHash())
}
