package dag

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
)

// dotEdge is a graph edge. In directed graphs, the direction of the
// edge is given from -> to, otherwise the edge is semantically
// unordered.
type dotEdge struct {
	x, y *dotNode
}

// From returns the from node of the edge.
func (e *dotEdge) From() graph.Node {
	return e.x
}

// To returns the to node of the edge.
func (e *dotEdge) To() graph.Node {
	return e.y
}

// ReversedEdge returns the edge reversal of the receiver
// if a reversal is valid for the data type.
// When a reversal is valid an edge of the same type as
// the receiver with nodes of the receiver swapped should
// be returned, otherwise the receiver should be returned
// unaltered.
func (e *dotEdge) ReversedEdge() graph.Edge {
	return nil
}

// dotNode is a graph node.
type dotNode struct {
	id      int64
	hash    hash.Event
	parents hash.Events
	frame   idx.Frame
	attrs   map[string]string
}

func newDotNode(id int64, e dag.Event) *dotNode {
	n := &dotNode{
		id:      id,
		hash:    e.ID(),
		parents: e.Parents(),
		attrs:   make(map[string]string, 10),
	}
	n.setAttr("label", n.hash.String())
	return n
}

func (n *dotNode) ID() int64 {
	return n.id
}

func (n *dotNode) Attributes() []encoding.Attribute {
	aa := make([]encoding.Attribute, 0, len(n.attrs))

	for k, v := range n.attrs {
		aa = append(aa,
			encoding.Attribute{
				Key:   k,
				Value: v,
			})
	}

	return aa
}

func (n *dotNode) setAttr(key, val string) {
	if val == "" {
		delete(n.attrs, key)
		return
	}
	n.attrs[key] = val
}

type dagNodes struct {
	data    chan *dotNode
	current *dotNode
}

// Reset returns the iterator to its start position.
func (nn *dagNodes) Reset() {
	panic("Not implemented yet")
}

// Next advances the iterator and returns whether
// the next call to the item method will return a
// non-nil item.
//
// Next should be called prior to any call to the
// iterator's item retrieval method after the
// iterator has been obtained or reset.
//
// The order of iteration is implementation
// dependent.
func (nn *dagNodes) Next() bool {
	nn.current = <-nn.data
	return nn.current != nil
}

// Node returns the current Node from the iterator.
func (nn *dagNodes) Node() graph.Node {
	return nn.current
}

// Len returns the number of items remaining in the
// iterator.
//
// If the number of items in the iterator is unknown,
// too large to materialize or too costly to calculate
// then Len may return a negative value.
// In this case the consuming function must be able
// to operate on the items of the iterator directly
// without materializing the items into a slice.
// The magnitude of a negative length has
// implementation-dependent semantics.
func (nn *dagNodes) Len() int {
	return -1
}
