package trie

import (
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/trie/internal"
)

/*
 * Marshal:
 */

// EncodeToBytes serializes node.
func EncodeToBytes(n node) ([]byte, error) {
	box := n.ToWire()
	return proto.Marshal(box)
}

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
 * Unmarshal:
 */

func mustDecodeNode(hash, buf []byte, cachegen uint16) node {
	n, err := decodeNode(hash, buf, cachegen)
	if err != nil {
		panic(fmt.Sprintf("node %x: %v", hash, err))
	}
	return n
}

func decodeNode(hash, buf []byte, cachegen uint16) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	var x internal.Node
	err := proto.Unmarshal(buf, &x)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	switch x.Kind.(type) {
	case *internal.Node_Short:
		n, err := decodeShort(hash, x.GetShort(), cachegen)
		return n, wrapError(err, "short")
	case *internal.Node_Full:
		n, err := decodeFull(hash, x.GetFull(), cachegen)
		return n, wrapError(err, "full")
	default:
		return nil, fmt.Errorf("unexpected node: %+v", x.Kind)
	}
}

func decodeShort(hash []byte, x *internal.ShortNode, cachegen uint16) (node, error) {
	flag := nodeFlag{hash: hash, gen: cachegen}
	key := compactToHex(x.Key)
	if hasTerm(key) {
		// value node
		if v, ok := x.Val.Kind.(*internal.Node_Value); ok {
			return &shortNode{key, valueNode(v.Value), flag}, nil
		} else {
			return nil, fmt.Errorf("invalid value node: %+v", x.Val.Kind)
		}
	}
	r, err := decodeRef(x.Val, cachegen)
	if err != nil {
		return nil, wrapError(err, "val")
	}
	return &shortNode{key, r, flag}, nil
}

func decodeFull(hash []byte, x *internal.FullNode, cachegen uint16) (*fullNode, error) {
	n := &fullNode{flags: nodeFlag{hash: hash, gen: cachegen}}
	for i := 0; i < 16; i++ {
		cld, err := decodeRef(x.Children[i], cachegen)
		if err != nil {
			return n, wrapError(err, fmt.Sprintf("[%d]", i))
		}
		n.Children[i] = cld
	}

	// value node
	if v, ok := x.Children[16].Kind.(*internal.Node_Value); ok {
		if len(v.Value) > 0 {
			n.Children[16] = valueNode(v.Value)
		}
	} else {
		panic(fmt.Errorf("unexpected node: %+v", x.Children[16].Kind))
	}

	return n, nil
}

func decodeRef(x *internal.Node, cachegen uint16) (node, error) {
	switch x.Kind.(type) {
	case *internal.Node_Hash:
		return hashNode(x.GetHash()), nil
	case *internal.Node_Value:
		if len(x.GetValue()) == 0 {
			return nil, nil
		}
		return valueNode(x.GetValue()), nil
	case *internal.Node_Short:
		return decodeShort(nil, x.GetShort(), cachegen)
	case *internal.Node_Full:
		return decodeFull(nil, x.GetFull(), cachegen)
	default:
		return nil, fmt.Errorf("unexpected node: %+v", x.Kind)
	}
}
