package dag

import (
	"fmt"
	"math"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagordering"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
	"github.com/Fantom-foundation/go-opera/utils/dag/dot"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

// readDagGraph read gossip.Store into inmem dot.Graph
func readDagGraph(gdb *gossip.Store, cfg integration.Configs, from, to idx.Epoch) *dot.Graph {
	// 0. Set gossip data:

	cdb := abft.NewMemStore()
	defer cdb.Close()
	// ApplyGenesis()
	cdb.SetEpochState(&abft.EpochState{
		Epoch: from,
	})
	cdb.SetLastDecidedState(&abft.LastDecidedState{
		LastDecidedFrame: abft.FirstFrame - 1,
	})

	dagIndexer := vecmt.NewIndex(panics("Vector clock"), cfg.VectorClock)
	orderer := abft.NewOrderer(
		cdb,
		&integration.GossipStoreAdapter{gdb},
		vecmt2dagidx.Wrap(dagIndexer),
		panics("Lachesis"),
		cfg.Lachesis)
	err := orderer.Bootstrap(abft.OrdererCallbacks{})
	if err != nil {
		panic(err)
	}

	// 1. Set dot.Graph data:

	g := dot.NewGraph("DOT")
	g.Set("clusterrank", "local")
	g.Set("compound", "true")
	g.Set("newrank", "true")
	g.Set("ranksep", "0.05")
	g.SetGlobalEdgeAttr("constraint", "true")

	var (
		nodes     = make(map[hash.Event]*dot.Node)
		subGraphs = make(map[idx.ValidatorID]*dot.SubGraph)
	)

	// 2. Set event processor data:

	var (
		epoch     idx.Epoch
		prevBS    *iblockproc.BlockState
		processed map[hash.Event]dag.Event
	)

	readRestoredAbftStore := func() {
		bs, _ := gdb.GetHistoryBlockEpochState(epoch)

		for f := idx.Frame(0); f <= cdb.GetLastDecidedFrame(); f++ {
			rr := cdb.GetFrameRoots(f)
			for _, r := range rr {
				n := nodes[r.ID]
				markAsRoot(n)
			}
		}

		if prevBS != nil {

			maxBlock := idx.Block(math.MaxUint64)
			if bs != nil {
				maxBlock = bs.LastBlock.Idx
			}

			for b := prevBS.LastBlock.Idx + 1; b <= maxBlock; b++ {
				block := gdb.GetBlock(b)
				if block == nil {
					break
				}
				n := nodes[block.Atropos]
				if n == nil {
					continue
				}
				markAsAtropos(n)
			}
		}

		prevBS = bs
	}

	resetToNewEpoch := func() {
		validators := gdb.GetHistoryEpochState(epoch).Validators

		for _, v := range validators.IDs() {
			if _, ok := subGraphs[v]; ok {
				continue
			}
			group := groupName(v)
			sg := dot.NewSubgraph(fmt.Sprintf("cluster%d", len(subGraphs)))
			sg.Set("label", group)
			sg.Set("sortv", fmt.Sprintf("%d", v))
			sg.Set("style", "dotted")

			pseudoNode := dot.NewNode(group)
			pseudoNode.Set("style", "invis")
			pseudoNode.Set("width", "0")
			sg.AddNode(pseudoNode)

			subGraphs[v] = sg
			g.AddSubgraph(sg)
		}

		processed = make(map[hash.Event]dag.Event, 1000)
		err := orderer.Reset(epoch, validators)
		if err != nil {
			panic(err)
		}
		dagIndexer.Reset(validators, memorydb.New(), func(id hash.Event) dag.Event {
			return gdb.GetEvent(id)
		})
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

				name := fmt.Sprintf("%s\n%d", e.ID().String(), e.Creator())
				n := dot.NewNode(name)
				sg := subGraphs[e.Creator()]
				sg.AddNode(n)
				nodes[e.ID()] = n

				for _, h := range e.Parents() {
					p := nodes[h]
					ref := dot.NewEdge(n, p)
					if processed[h].Creator() == e.Creator() {
						sg.AddEdge(ref)
					} else {
						g.AddEdge(ref)
					}
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

	// 3. Iterate over events:

	gdb.ForEachEvent(from, func(e *inter.EventPayload) bool {
		// current epoch is finished, so process accumulated events
		if epoch < e.Epoch() {
			readRestoredAbftStore()

			epoch = e.Epoch()
			// break after last epoch:
			if to >= from && epoch > to {
				return false
			}

			resetToNewEpoch()
		}

		buffer.PushEvent(e, "")

		return true
	})
	epoch++
	readRestoredAbftStore()

	// 4. Result

	// NOTE: github.com/tmc/dot renders subgraphs not in the ordering that specified
	//   so we introduce pseudo nodes and edges to work around
	groups := make([]string, 0, len(subGraphs))
	for v := range subGraphs {
		groups = append(groups, groupName(v))
	}
	g.SameRank([]string{
		"\"" + strings.Join(groups, `" -> "`) + "\" [style = invis, constraint = true];",
	})

	return g
}

func panics(name string) func(error) {
	return func(err error) {
		log.Crit(fmt.Sprintf("%s error", name), "err", err)
	}
}

func markAsRoot(n *dot.Node) {
	// n.setAttr("xlabel", "root")
	n.Set("style", "filled")
	n.Set("fillcolor", "#FFFF00")
}

func markAsAtropos(n *dot.Node) {
	// n.setAttr("xlabel", "atropos")
	n.Set("style", "filled")
	n.Set("fillcolor", "#FF0000")
}

func groupName(v idx.ValidatorID) string {
	return fmt.Sprintf("host-%d", v)
}
