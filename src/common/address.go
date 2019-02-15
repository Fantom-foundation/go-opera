package common

var (
	// NodeNameDict is an optional dictionary to make node address human readable in log.
	NodeNameDict = make(map[Address]string)
)

// Address is a unique identificator of Node.
// It is a hash of node's PubKey.
type Address Hash

// Bytes returns value as byte slice.
func (a *Address) Bytes() []byte {
	return (*Hash)(a).Bytes()
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

func FakeAddress() Address {
	return Address(FakeHash())
}
