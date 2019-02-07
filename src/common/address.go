package common

import (
	"math/rand"
)

// Address is a unique identificator of Node.
// It is a hash of node's PubKey.
type Address Hash

// Bytes returns value as byte slice.
func (a *Address) Bytes() []byte {
	return (*Hash)(a).Bytes()
}

// String returns value as hex string.
func (a *Address) String() string {
	return (*Hash)(a).String()
}

/*
 * Utils:
 */

func FakeAddress() (a Address) {
	_, err := rand.Read(a[:])
	if err != nil {
		panic(err)
	}
	return
}
