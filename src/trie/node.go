package trie

import (
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/trie/internal"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
	fstring(string) string
	cache() (hashNode, bool)
	canUnload(cachegen, cachelimit uint16) bool

	ToWire() *internal.Node
}

type (
	fullNode struct {
		Children [17]node // Actual trie node data to encode/decode (needs custom encoder)
		flags    nodeFlag
	}
	shortNode struct {
		Key   []byte
		Val   node
		flags nodeFlag
	}
	hashNode  []byte
	valueNode []byte
)

// nilValueNode is used when collapsing internal trie nodes for hashing, since
// unset children need to serialize correctly.
var nilValueNode = valueNode(nil)

func (n *fullNode) copy() *fullNode   { _copy := *n; return &_copy }
func (n *shortNode) copy() *shortNode { _copy := *n; return &_copy }

// nodeFlag contains caching-related metadata about a node.
type nodeFlag struct {
	hash  hashNode // cached hash of the node (may be nil)
	gen   uint16   // cache generation counter
	dirty bool     // whether the node has changes that must be written to the database
}

// canUnload tells whether a node can be unloaded.
func (n *nodeFlag) canUnload(cachegen, cachelimit uint16) bool {
	return !n.dirty && cachegen-n.gen >= cachelimit
}

func (n *fullNode) canUnload(gen, limit uint16) bool  { return n.flags.canUnload(gen, limit) }
func (n *shortNode) canUnload(gen, limit uint16) bool { return n.flags.canUnload(gen, limit) }
func (n hashNode) canUnload(uint16, uint16) bool      { return false }
func (n valueNode) canUnload(uint16, uint16) bool     { return false }

func (n *fullNode) cache() (hashNode, bool)  { return n.flags.hash, n.flags.dirty }
func (n *shortNode) cache() (hashNode, bool) { return n.flags.hash, n.flags.dirty }
func (n hashNode) cache() (hashNode, bool)   { return nil, true }
func (n valueNode) cache() (hashNode, bool)  { return nil, true }

// Pretty printing.
func (n *fullNode) String() string  { return n.fstring("") }
func (n *shortNode) String() string { return n.fstring("") }
func (n hashNode) String() string   { return n.fstring("") }
func (n valueNode) String() string  { return n.fstring("") }

func (n *fullNode) fstring(ind string) string {
	resp := fmt.Sprintf("[\n%s  ", ind)
	for i, node := range &n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil> ", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fstring(ind+"  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}
func (n *shortNode) fstring(ind string) string {
	return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fstring(ind+"  "))
}
func (n hashNode) fstring(ind string) string {
	return fmt.Sprintf("<%x> ", []byte(n))
}
func (n valueNode) fstring(ind string) string {
	return fmt.Sprintf("%x ", []byte(n))
}

func mustDecodeNode(hash, buf []byte, cachegen uint16) node {
	n, err := decodeNode(hash, buf, cachegen)
	if err != nil {
		panic(fmt.Sprintf("node %x: %v", hash, err))
	}
	return n
}

// decodeNode parses the protobuf encoding of a trie node.
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
		if len(x.GetValue()) > 0 {
			return valueNode(x.GetValue()), nil
		} else {
			return nil, nil
		}
	case *internal.Node_Short:
		return decodeShort(nil, x.GetShort(), cachegen)
	case *internal.Node_Full:
		return decodeFull(nil, x.GetFull(), cachegen)
	default:
		return nil, fmt.Errorf("unexpected node: %+v", x.Kind)
	}
}

// wraps a decoding error with information about the path to the
// invalid child node (for debugging encoding issues).
type decodeError struct {
	what  error
	stack []string
}

func wrapError(err error, ctx string) error {
	if err == nil {
		return nil
	}
	if decErr, ok := err.(*decodeError); ok {
		decErr.stack = append(decErr.stack, ctx)
		return decErr
	}
	return &decodeError{err, []string{ctx}}
}

func (err *decodeError) Error() string {
	return fmt.Sprintf("%v (decode path: %s)", err.what, strings.Join(err.stack, "<-"))
}
