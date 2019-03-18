package trie

import (
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/trie/internal"
)

// ToWire converts node to wire.
func (n *shortNode) ToWire() *internal.Node {
	return &internal.Node{
		Kind: &internal.Node_Short{
			Short: &internal.ShortNode{
				Key: n.Key,
				Val: n.Val.ToWire(),
			},
		},
	}
}

// ToWire converts node to wire.
func (n *rawShortNode) ToWire() *internal.Node {
	wrapper := shortNode{
		Key: n.Key,
		Val: n.Val,
	}
	return wrapper.ToWire()
}

// ToWire converts node to wire.
func (n *fullNode) ToWire() *internal.Node {
	nodes := make([]*internal.Node, len(&n.Children))
	for i, child := range &n.Children {
		if child != nil {
			nodes[i] = child.ToWire()
		} else {
			nodes[i] = &internal.Node{
				Kind: &internal.Node_Value{
					Value: []byte(nil),
				},
			}
		}
	}

	return &internal.Node{
		Kind: &internal.Node_Full{
			Full: &internal.FullNode{
				Children: nodes,
			},
		},
	}
}

// ToWire converts node to wire.
func (n rawFullNode) ToWire() *internal.Node {
	wrapper := fullNode{
		Children: n,
	}
	return wrapper.ToWire()
}

// ToWire converts node to wire.
func (n rawNode) ToWire() *internal.Node {
	panic("this should never end up in a live trie")
}

// ToWire converts node to wire.
func (n hashNode) ToWire() *internal.Node {
	return &internal.Node{
		Kind: &internal.Node_Hash{
			Hash: []byte(n),
		},
	}
}

// ToWire converts node to wire.
func (n valueNode) ToWire() *internal.Node {
	return &internal.Node{
		Kind: &internal.Node_Value{
			Value: []byte(n),
		},
	}
}

/*
 * Utils:
 */

// EncodeToBytes serializes node.
func EncodeToBytes(n node) ([]byte, error) {
	box := n.ToWire()
	return proto.Marshal(box)
}
