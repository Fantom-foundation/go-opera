package posposet

// Node is a event's author.
// TODO: move to new src/node/.
type Node struct {
	ID     Address
	PubKey PublicKey

	key *PrivateKey
}

// NewNode creates Node instance.
func NewNode(pk PublicKey) *Node {
	return &Node{
		ID:     AddressOf(pk),
		PubKey: pk,
	}
}
