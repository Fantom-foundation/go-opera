package posnode

type Node struct {
	consensus Consensus
}

func New(c Consensus) *Node {
	return &Node{
		consensus: c,
	}
}
