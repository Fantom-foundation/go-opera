package dag

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"gonum.org/v1/gonum/graph"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter"
)

// dagReader implements dot.Graph over gossip.Store
type dagReader struct {
	db        *gossip.Store
	epochFrom idx.Epoch
	epochTo   idx.Epoch
}

func (g *dagReader) DOTID() string {
	return "DAG"
}

// Node returns the node with the given ID if it exists
// in the graph, and nil otherwise.
func (g *dagReader) Node(id int64) graph.Node {
	e := g.db.GetEvent(id2event(id))
	return &dagNode{
		id:    event2id(e.ID()),
		event: e,
	}
}

// Nodes returns all the nodes in the graph.
//
// Nodes must not return nil.
func (g *dagReader) Nodes() graph.Nodes {
	nn := &dagNodes{
		data: make(chan *inter.Event),
	}

	go func() {
		defer close(nn.data)
		g.db.ForEachEvent(g.epochFrom, func(e *inter.EventPayload) bool {
			if e.Epoch() > g.epochTo {
				return false
			}

			nn.data <- &e.Event
			return true
		})
	}()

	return nn
}

// From returns all nodes that can be reached directly
// from the node with the given ID.
//
// From must not return nil.
func (g *dagReader) From(id int64) graph.Nodes {
	nn := &dagNodes{
		data: make(chan *inter.Event),
	}

	x := g.Node(id).(*dagNode).event
	go func() {
		defer close(nn.data)
		for _, p := range x.Parents() {
			n := g.Node(event2id(p))
			nn.data <- n.(*dagNode).event
		}
	}()

	return nn
}

// HasEdgeBetween returns whether an edge exists between
// nodes with IDs xid and yid without considering direction.
func (g *dagReader) HasEdgeBetween(xid, yid int64) bool {
	x := g.Node(xid).(*dagNode).event
	y := g.Node(yid).(*dagNode).event

	for _, p := range x.Parents() {
		if p == y.ID() {
			return true
		}
	}
	for _, p := range y.Parents() {
		if p == x.ID() {
			return true
		}
	}

	return false
}

// Edge returns the edge from u to v, with IDs uid and vid,
// if such an edge exists and nil otherwise. The node v
// must be directly reachable from u as defined by the
// From method.
func (g *dagReader) Edge(uid, vid int64) graph.Edge {
	u := g.Node(uid).(*dagNode)
	v := g.Node(vid).(*dagNode)

	for _, p := range u.event.Parents() {
		if p == v.event.ID() {
			return &dagEdge{
				x: u,
				y: v,
			}
		}
	}

	return nil
}

// --

var (
	id2hash = make(map[int64]hash.Event)
)

func id2event(id int64) hash.Event {
	return id2hash[id]
}

func event2id(e hash.Event) int64 {
	// NOTE: possible collision
	var id int64
	for i := 0; i < 8; i++ {
		id += int64(e[8+i] << (8 * i))
	}

	id2hash[id] = e

	return id
}
