package hash

var (
	// NodeNameDict is an optional dictionary to make node address human readable in log.
	NodeNameDict = make(map[Peer]string)
)

// Peer is a unique peer identificator.
// It is a hash of peer's PubKey.
type Peer Hash

// Bytes returns value as byte slice.
func (a *Peer) Bytes() []byte {
	return (*Hash)(a).Bytes()
}

// BytesToPeer converts bytes to peer id.
// If b is larger than len(h), b will be cropped from the left.
func BytesToPeer(b []byte) Peer {
	return Peer(FromBytes(b))
}

// String returns human readable string representation.
func (a *Peer) String() string {
	if name, ok := NodeNameDict[*a]; ok {
		return name
	}
	return (*Hash)(a).ShortString()
}

/*
 * Utils:
 */

// FakePeer generates random fake peer id for testing purpose.
func FakePeer() Peer {
	return Peer(FakeHash())
}
