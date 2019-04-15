package hash

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

var (
	// NodeNameDict is an optional dictionary to make node address human readable in log.
	NodeNameDict = make(map[Peer]string)
)

// Peer is a unique peer identificator.
// It is a hash of peer's PubKey.
type Peer Hash

// EmptyPeer is empty peer identificator.
var EmptyPeer = Peer{}

// Bytes returns value as byte slice.
func (p *Peer) Bytes() []byte {
	return (*Hash)(p).Bytes()
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
func PeerOfPubkey(pub *common.PublicKey) Peer {
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
