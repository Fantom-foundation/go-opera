package posposet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

func TestPoset(t *testing.T) {
	logger.SetTestMode(t)

	const posetCount = 3 // last will be restored

	nodes, events := inter.GenEventsByNode(5, 99, 3)

	posets := make([]*Poset, 0, posetCount)
	inputs := make([]*EventStore, 0, posetCount)

	makePoset := func(i int) *Store {
		poset, store, input := FakePoset(nodes)
		n := i % len(nodes)
		poset.SetName(nodes[n].String())
		store.SetName(nodes[n].String())
		poset.Start()
		posets = append(posets, poset)
		inputs = append(inputs, input)
		return store
	}

	for i := 0; i < posetCount-1; i++ {
		_ = makePoset(i)
	}

	t.Run("Multiple start", func(t *testing.T) {
		posets[0].Stop()
		posets[0].Start()
		posets[0].Start()
	})

	t.Run("Push unordered events", func(t *testing.T) {
		// first all events from one node
		for i := 0; i < len(posets); i++ {
			n := i % len(nodes)
			ee := events[nodes[n]]
			for _, e := range ee {
				inputs[i].SetEvent(e)
				posets[i].PushEventSync(e.Hash())
			}
		}
		// second all events from others
		for i := 0; i < len(posets); i++ {
			for n := range nodes {
				if n == i%len(nodes) {
					continue
				}
				ee := events[nodes[n]]
				for _, e := range ee {
					inputs[i].SetEvent(e)
					posets[i].PushEventSync(e.Hash())
				}
			}
		}
	})

	t.Run("Restore", func(t *testing.T) {
		i := posetCount - 1
		store := makePoset(i)

		all := inter.Events{}
		for n := range nodes {
			ee := events[nodes[n]]
			for _, e := range ee {
				all = append(all, e)
			}
		}

		for x, e := range all {
			if x == len(all)/2 {
				// restore
				posets[i].Stop()
				restored := New(store, inputs[i])
				n := i % len(nodes)
				restored.SetName("restored_" + nodes[n].String())
				store.SetName("restored_" + nodes[n].String())
				restored.Bootstrap()
				MakeOrderedInput(restored)
				posets[i] = restored
			}

			inputs[i].SetEvent(e)
			posets[i].PushEventSync(e.Hash())
		}
	})

	t.Run("Check consensus", func(t *testing.T) {
		assertar := assert.New(t)
		for i := 0; i < len(posets)-1; i++ {
			p0 := posets[i]
			st0 := p0.store.GetCheckpoint()
			t.Logf("Compare poset%d: SFrame %d, Block %d", i, st0.SuperFrameN, st0.LastBlockN)
			for j := i + 1; j < len(posets); j++ {
				p1 := posets[j]
				st1 := p1.store.GetCheckpoint()
				t.Logf("with poset%d: SFrame %d, Block %d", j, st1.SuperFrameN, st1.LastBlockN)

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

	t.Run("Multiple stop", func(t *testing.T) {
		posets[0].Stop()
		posets[0].Stop()
	})

	for i := 0; i < len(posets); i++ {
		posets[i].Stop()
	}
}
