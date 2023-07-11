package dag

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"gonum.org/v1/gonum/graph"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter"
)

type dagLoader struct {
	refs  []hash.Event
	nodes map[hash.Event]*dagNode
}

func newDagLoader(db *gossip.Store, from, to idx.Epoch) *dagLoader {
	g := &dagLoader{
		refs:  make([]hash.Event, 0, 2000000),
		nodes: make(map[hash.Event]*dagNode),
	}

	db.ForEachEvent(from, func(e *inter.EventPayload) bool {
		if to >= from && e.Epoch() > to {
			return false
		}

		id := len(g.refs)
		g.refs = append(g.refs, e.ID())
		g.nodes[e.ID()] = &dagNode{
			id:      int64(id),
			hash:    e.ID(),
			parents: e.Parents(),
		}

		return true
	})

	db.ForEachBlock(func(index idx.Block, block *inter.Block) {
		node, exists := g.nodes[block.Atropos]
		if exists {
			node.isAtropos = true
		}
	})

	return g
}

func (g *dagLoader) DOTID() string {
	return "DAG"
}

// Node returns the node with the given ID if it exists
// in the graph, and nil otherwise.
func (g *dagLoader) Node(id int64) graph.Node {
	hash := g.refs[id]
	return g.nodes[hash]
}

// Nodes returns all the nodes in the graph.
//
// Nodes must not return nil.
func (g *dagLoader) Nodes() graph.Nodes {
	nn := &dagNodes{
		data: make(chan *dagNode),
	}

	go func() {
		defer close(nn.data)

		for _, e := range g.nodes {
			nn.data <- e
		}
	}()

	return nn
}

// From returns all nodes that can be reached directly
// from the node with the given ID.
//
// From must not return nil.
func (g *dagLoader) From(id int64) graph.Nodes {
	nn := &dagNodes{
		data: make(chan *dagNode),
	}

	h := g.refs[id]
	x := g.nodes[h]
	go func() {
		defer close(nn.data)
		for _, p := range x.parents {
			n := g.nodes[p]
			nn.data <- n
		}
	}()

	return nn
}

// To returns all nodes that can reach directly
// to the node with the given ID.
//
// To must not return nil.
func (g *dagLoader) To(id int64) graph.Nodes {
	nn := &dagNodes{
		data: make(chan *dagNode),
	}
	close(nn.data)
	return nn
}

// HasEdgeBetween returns whether an edge exists between
// nodes with IDs xid and yid without considering direction.
func (g *dagLoader) HasEdgeBetween(xid, yid int64) bool {
	x := g.nodes[g.refs[xid]]
	y := g.nodes[g.refs[yid]]

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
func (g *dagLoader) HasEdgeFromTo(uid, vid int64) bool {
	u := g.nodes[g.refs[uid]]
	v := g.nodes[g.refs[vid]]

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
func (g *dagLoader) Edge(uid, vid int64) graph.Edge {
	u := g.nodes[g.refs[uid]]
	v := g.nodes[g.refs[vid]]

	for _, p := range u.parents {
		if p == v.hash {
			return &dagEdge{
				x: u,
				y: v,
			}
		}
	}

	return nil
}
