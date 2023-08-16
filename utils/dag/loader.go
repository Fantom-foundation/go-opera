package dag

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagordering"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/ethereum/go-ethereum/log"
	"gonum.org/v1/gonum/graph"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

type dagLoader struct {
	refs  []hash.Event
	nodes map[hash.Event]*dagNode
}

func newDagLoader(gdb *gossip.Store, cfg integration.Configs, from, to idx.Epoch) *dagLoader {
	g := &dagLoader{
		refs:  make([]hash.Event, 0, 2000000),
		nodes: make(map[hash.Event]*dagNode),
	}

	store := abft.NewMemStore()
	defer store.Close()
	// ApplyGenesis()
	store.SetEpochState(&abft.EpochState{
		Epoch: from,
	})
	store.SetLastDecidedState(&abft.LastDecidedState{
		LastDecidedFrame: abft.FirstFrame - 1,
	})

	dagIndexer := vecmt.NewIndex(panics("Vector clock"), cfg.VectorClock)
	orderer := abft.NewOrderer(
		store,
		&integration.GossipStoreAdapter{gdb},
		vecmt2dagidx.Wrap(dagIndexer),
		panics("Lachesis"),
		cfg.Lachesis)

	var (
		epoch     idx.Epoch
		prevBS    *iblockproc.BlockState
		processed map[hash.Event]dag.Event
	)
	err := orderer.Bootstrap(abft.OrdererCallbacks{
		ApplyAtropos: func(decidedFrame idx.Frame, atropos hash.Event) (sealEpoch *pos.Validators) {
			return nil
		},
	})
	if err != nil {
		panic(err)
	}
	buffer := dagordering.New(
		cfg.Opera.Protocol.DagProcessor.EventsBufferLimit,
		dagordering.Callback{
			Process: func(e dag.Event) error {
				processed[e.ID()] = e
				err = dagIndexer.Add(e)
				if err != nil {
					panic(err)
				}
				dagIndexer.Flush()
				orderer.Process(e)

				id := len(g.refs)
				g.refs = append(g.refs, e.ID())
				g.nodes[e.ID()] = &dagNode{
					id:      int64(id),
					hash:    e.ID(),
					parents: e.Parents(),
					frame:   e.Frame(),
				}
				return nil
			},
			Released: func(e dag.Event, peer string, err error) {
				if err != nil {
					panic(err)
				}
			},
			Get: func(id hash.Event) dag.Event {
				return processed[id]
			},
			Exists: func(id hash.Event) bool {
				_, ok := processed[id]
				return ok
			},
		})

	gdb.ForEachEvent(from, func(e *inter.EventPayload) bool {
		// current epoch is finished, so process accumulated events
		if epoch < e.Epoch() {
			epoch = e.Epoch()
			bs, es := gdb.GetHistoryBlockEpochState(epoch)

			// data from restored abft store:

			for f := idx.Frame(0); f <= store.GetLastDecidedFrame(); f++ {
				rr := store.GetFrameRoots(f)
				for _, r := range rr {
					g.nodes[r.ID].isRoot = true
				}
			}

			if prevBS != nil {
				for n := prevBS.LastBlock.Idx + 1; n <= bs.LastBlock.Idx; n++ {
					block := gdb.GetBlock(n)
					node, exists := g.nodes[block.Atropos]
					if exists {
						node.isAtropos = true
					}
				}
			}

			// break after last epoch:
			if to >= from && epoch > to {
				return false
			}

			// reset to new epoch:

			prevBS = bs
			processed = make(map[hash.Event]dag.Event, 1000)
			err := orderer.Reset(epoch, es.Validators)
			if err != nil {
				panic(err)
			}
			dagIndexer.Reset(es.Validators, memorydb.New(), func(id hash.Event) dag.Event {
				return gdb.GetEvent(id)
			})
		}

		// process epoch's event
		buffer.PushEvent(e, "")

		return true
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

func panics(name string) func(error) {
	return func(err error) {
		log.Crit(fmt.Sprintf("%s error", name), "err", err)
	}
}
