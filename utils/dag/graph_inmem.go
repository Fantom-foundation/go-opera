package dag

import (
	"fmt"
	"math"
	"sort"
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

	graph := dot.NewGraph("DOT")
	graph.Set("compound", "true")
	graph.SetGlobalEdgeAttr("constraint", "true")
	var (
		clusters  idx.ValidatorID
		g         *dot.SubGraph // epoch sub
		subGraphs map[idx.ValidatorID]*dot.SubGraph
		emitters  []string
		nodes     map[hash.Event]*dot.Node
	)

	// 2. Set event processor data:

	var (
		epoch     idx.Epoch
		processed map[hash.Event]dag.Event
	)

	finishCurrentEpoch := func() {
		if epoch == 0 {
			return // nothing to finish
		}

		for f := idx.Frame(0); f <= cdb.GetLastDecidedFrame(); f++ {
			rr := cdb.GetFrameRoots(f)
			for _, r := range rr {
				n := nodes[r.ID]
				markAsRoot(n)
			}
		}

		bs0, _ := gdb.GetHistoryBlockEpochState(epoch)
		blockFrom := bs0.LastBlock.Idx + 1

		bs1, _ := gdb.GetHistoryBlockEpochState(epoch + 1)
		blockTo := idx.Block(math.MaxUint64)
		if bs1 != nil {
			blockTo = bs1.LastBlock.Idx
		}

		for b := blockFrom; b <= blockTo; b++ {
			block := gdb.GetBlock(b)
			if block == nil {
				break
			}
			n := nodes[block.Atropos]
			markAsAtropos(n)
		}

		// NOTE: github.com/tmc/dot renders subgraphs not in the ordering that specified
		//   so we introduce pseudo nodes and edges to work around
		g.SameRank([]string{
			"\"" + strings.Join(emitters, `" -> "`) + "\" [style = invis, constraint = true];",
		})
	}

	resetToNewEpoch := func() {
		name := fmt.Sprintf("epoch-%d", epoch)
		g = dot.NewSubgraph(name)
		g.Set("label", name)
		g.Set("style", "dotted")
		g.Set("compound", "true")
		g.Set("clusterrank", "local")
		g.Set("newrank", "true")
		g.Set("ranksep", "0.05")
		graph.AddSubgraph(g)

		validators := gdb.GetHistoryEpochState(epoch).Validators

		sortedIDs := make([]idx.ValidatorID, validators.Len())
		copy(sortedIDs, validators.IDs())
		sort.Slice(sortedIDs, func(i, j int) bool {
			return sortedIDs[i] < sortedIDs[j]
		})
		subGraphs = make(map[idx.ValidatorID]*dot.SubGraph, len(sortedIDs))
		emitters = make([]string, 0, len(sortedIDs))
		var maxID idx.ValidatorID
		for _, v := range sortedIDs {
			if maxID < v {
				maxID = v
			}
			emitter := fmt.Sprintf("emitter-%d", clusters+v)
			emitters = append(emitters, emitter)
			sg := dot.NewSubgraph(fmt.Sprintf("cluster%d", clusters+v))
			sg.Set("label", fmt.Sprintf("emitter-%d", v))
			sg.Set("sortv", fmt.Sprintf("%d", clusters+v))
			sg.Set("style", "dotted")
			subGraphs[v] = sg
			g.AddSubgraph(sg)

			pseudoNode := dot.NewNode(emitter)
			pseudoNode.Set("style", "invis")
			pseudoNode.Set("width", "0")
			sg.AddNode(pseudoNode)
		}
		clusters += maxID

		nodes = make(map[hash.Event]*dot.Node)
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

				name := e.ID().String()
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
			// break after last epoch:
			if to >= from && e.Epoch() > to {
				return false
			}
			finishCurrentEpoch()
			epoch = e.Epoch()
			resetToNewEpoch()
		}

		buffer.PushEvent(e, "")
		return true
	})
	finishCurrentEpoch()

	// 4. Result

	return graph
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
