package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPoset(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const posetCount = 3

	nodes := inter.GenNodes(5)

	posets := make([]*BufferedPoset, 0, posetCount)
	inputs := make([]*EventStore, 0, posetCount)

	makePoset := func(i int) *Store {
		poset, store, input := FakePoset(nodes)
		n := i % len(nodes)
		poset.SetName(nodes[n].String())
		store.SetName(nodes[n].String())
		posets = append(posets, poset)
		inputs = append(inputs, input)
		return store
	}

	for i := 0; i < posetCount-1; i++ {
		_ = makePoset(i)
	}

	// create events on poset0
	unordered := inter.ForEachRandEvent(nodes, int(SuperFrameLen)-1, 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			inputs[0].SetEvent(e)
			assertar.NoError(posets[0].ProcessEvent(e))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			return posets[0].Prepare(e)
		},
	})

	pushedAll := false
	t.Run("Push unordered events", func(t *testing.T) {
		// first all events from one node
		for i := 1; i < len(posets); i++ {
			n := i % len(nodes)
			ee := unordered[nodes[n]]
			for _, e := range ee {
				if e.Epoch != 1 {
					continue
				}
				inputs[i].SetEvent(e)
				posets[i].PushToBuffer(e)
			}
		}
		// second all events from others
		for i := 1; i < len(posets); i++ {
			for n := range nodes {
				if n == i%len(nodes) {
					continue
				}
				ee := unordered[nodes[n]]
				for _, e := range ee {
					if e.Epoch != 1 {
						continue
					}
					inputs[i].SetEvent(e)
					posets[i].PushToBuffer(e)
				}
			}
		}
		pushedAll = true
	})

	t.Run("Check consensus", func(t *testing.T) {
		for i := 0; i < len(posets)-1; i++ {
			p0 := posets[i]
			st0 := p0.store.GetCheckpoint()
			ep0 := p0.store.GetSuperFrame()
			t.Logf("Compare poset%d: SFrame %d, Block %d", i, ep0.SuperFrameN, st0.LastBlockN)
			for j := i + 1; j < len(posets); j++ {
				p1 := posets[j]
				st1 := p1.store.GetCheckpoint()
				t.Logf("with poset%d: SFrame %d, Block %d", j, ep0.SuperFrameN, st1.LastBlockN)

				// compare state on p0/p1
				if pushedAll {
					assertar.Equal(*posets[j].checkpoint, *posets[i].checkpoint)
					assertar.Equal(posets[j].superFrame, posets[i].superFrame)
				}

				both := p0.LastBlockN
				if both > p1.LastBlockN {
					both = p1.LastBlockN
				}

				var failAt idx.Block
				for b := idx.Block(1); b <= both; b++ {
					if !assertar.Equal(
						p0.store.GetBlock(b).Events, p1.store.GetBlock(b).Events,
						"block %d", b) {
						failAt = b
						break
					}
				}
				if failAt == 0 {
					continue
				}

				scheme0, err := inter.DAGtoASCIIscheme(p0.EventsTillBlock(failAt))
				if err != nil {
					t.Fatal(err)
				}

				scheme1, err := inter.DAGtoASCIIscheme(p1.EventsTillBlock(failAt))
				if err != nil {
					t.Fatal(err)
				}

				DAGs := utils.TextColumns(scheme0, scheme1)
				t.Log(DAGs)

				return
			}
		}
	})
}
