package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

// Node is a event's author.
// TODO: move to new src/node/.
type Node struct {
	ID     common.Address
	PubKey common.PublicKey

	key *common.PrivateKey
}

// NewNode creates Node instance.
func NewNode(pk common.PublicKey) *Node {
	return &Node{
		ID:     crypto.AddressOf(pk),
		PubKey: pk,
	}
}
