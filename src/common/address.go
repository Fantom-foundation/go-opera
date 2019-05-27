package common

var (
	// NodeNameDict is an optional dictionary to make node address human readable in log.
	// TODO FIXIT: NodeNameDict is not populated.
	NodeNameDict = make(map[Address]string)
)

// Address is a unique identifier of Node.
// It is a hash of node's PubKey.
type Address Hash

// Bytes returns value as byte slice.
func (a *Address) Bytes() []byte {
	return (*Hash)(a).Bytes()
}

// BytesToAddress sets b to address.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	return Address(BytesToHash(b))
}

// String returns human readable string representation.
func (a *Address) String() string {
	if name, ok := NodeNameDict[*a]; ok {
		return name
	}
	return (*Hash)(a).ShortString()
}

/*
 * Utils:
 */

// FakeAddress generates random fake address for testing purpose.
func FakeAddress() Address {
	return Address(FakeHash())
}
