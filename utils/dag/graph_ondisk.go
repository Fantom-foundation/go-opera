package dag

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"gonum.org/v1/gonum/graph"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter"
)

// dagReader implements dot.Graph over gossip.Store
type graphOnDisk struct {
	db        *gossip.Store
	epochFrom idx.Epoch
	epochTo   idx.Epoch
}

func (g *graphOnDisk) DOTID() string {
	return "DAG"
}

// Node returns the node with the given ID if it exists
// in the graph, and nil otherwise.
func (g *graphOnDisk) Node(id int64) graph.Node {
	e := g.db.GetEvent(id2event(id))
	return newDotNode(id, e)
}

// Nodes returns all the nodes in the graph.
//
// Nodes must not return nil.
func (g *graphOnDisk) Nodes() graph.Nodes {
	nn := &dagNodes{
		data: make(chan *dotNode),
	}

	go func() {
		defer close(nn.data)
		g.db.ForEachEvent(g.epochFrom, func(e *inter.EventPayload) bool {
			if g.epochTo >= g.epochFrom && e.Epoch() > g.epochTo {
				return false
			}

			id := event2id(e.ID())
			nn.data <- newDotNode(id, &e.Event)
			return true
		})
	}()

	return nn
}

// From returns all nodes that can be reached directly
// from the node with the given ID.
//
// From must not return nil.
func (g *graphOnDisk) From(id int64) graph.Nodes {
	nn := &dagNodes{
		data: make(chan *dotNode),
	}

	x := g.Node(id).(*dotNode)
	go func() {
		defer close(nn.data)
		for _, p := range x.parents {
			n := g.Node(event2id(p))
			nn.data <- n.(*dotNode)
		}
	}()

	return nn
}

// To returns all nodes that can reach directly
// to the node with the given ID.
//
// To must not return nil.
func (g *graphOnDisk) To(id int64) graph.Nodes {
	nn := &dagNodes{
		data: make(chan *dotNode),
	}
	close(nn.data)
	return nn
}

// HasEdgeBetween returns whether an edge exists between
// nodes with IDs xid and yid without considering direction.
func (g *graphOnDisk) HasEdgeBetween(xid, yid int64) bool {
	x := g.Node(xid).(*dotNode)
	y := g.Node(yid).(*dotNode)

	for _, p := range x.parents {
		if p == y.hash {
			return true
		}
	}
	for _, p := range y.parents {
		if p == x.hash {
			return true
		}
	}

	return false
}

// HasEdgeFromTo returns whether an edge exists
// in the graph from u to v with IDs uid and vid.
func (g *graphOnDisk) HasEdgeFromTo(uid, vid int64) bool {
	u := g.Node(uid).(*dotNode)
	v := g.Node(vid).(*dotNode)

	for _, p := range u.parents {
		if p == v.hash {
			return true
		}
	}

	return false
}

// Edge returns the edge from u to v, with IDs uid and vid,
// if such an edge exists and nil otherwise. The node v
// must be directly reachable from u as defined by the
// From method.
func (g *graphOnDisk) Edge(uid, vid int64) graph.Edge {
	u := g.Node(uid).(*dotNode)
	v := g.Node(vid).(*dotNode)

	for _, p := range u.parents {
		if p == v.hash {
			return &dotEdge{
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
