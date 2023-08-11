package dag

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
)

// Edge is a graph edge. In directed graphs, the direction of the
// edge is given from -> to, otherwise the edge is semantically
// unordered.
type dagEdge struct {
	x, y *dagNode
}

// From returns the from node of the edge.
func (e *dagEdge) From() graph.Node {
	return e.x
}

// To returns the to node of the edge.
func (e *dagEdge) To() graph.Node {
	return e.y
}

// ReversedEdge returns the edge reversal of the receiver
// if a reversal is valid for the data type.
// When a reversal is valid an edge of the same type as
// the receiver with nodes of the receiver swapped should
// be returned, otherwise the receiver should be returned
// unaltered.
func (e *dagEdge) ReversedEdge() graph.Edge {
	return nil
}

type dagNode struct {
	id        int64
	hash      hash.Event
	parents   hash.Events
	isRoot    bool
	isAtropos bool
}

func (n *dagNode) ID() int64 {
	return n.id
}

func (n *dagNode) Attributes() []encoding.Attribute {
	aa := []encoding.Attribute{
		encoding.Attribute{
			Key:   "label",
			Value: n.hash.String(),
		},
	}

	if n.isRoot {
		aa = append(aa,
			encoding.Attribute{
				Key:   "role",
				Value: "Root",
			},
		)
	}

	if n.isAtropos {
		aa = append(aa,
			encoding.Attribute{
				Key:   "xlabel",
				Value: "Atropos",
			},
		)
	}

	return aa
}

type dagNodes struct {
	data    chan *dagNode
	current *dagNode
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
