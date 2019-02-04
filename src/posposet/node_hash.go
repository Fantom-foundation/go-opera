package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

// Address is a unique identificator of Node.
// It is a hash of node's PubKey.
type Address common.Hash

// AddressOf calcs hash of the PublicKey.
func AddressOf(pk PublicKey) Address {
	return Address(crypto.Keccak256Hash(pk.Bytes()))
}

// String returns value as hex string.
func (a *Address) String() string {
	return (*common.Hash)(a).String()
}
