package poset

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestPoset(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const posetCount = 3
	nodes := inter.GenNodes(5)

	posets := make([]*ExtendedPoset, 0, posetCount)
	inputs := make([]*EventStore, 0, posetCount)
	for i := 0; i < posetCount-1; i++ {
		poset, store, input := FakePoset(nodes)
		n := i % len(nodes)
		poset.SetName(nodes[n].String())
		store.SetName(nodes[n].String())
		posets = append(posets, poset)
		inputs = append(inputs, input)
	}

	// create events on poset0
	var ordered inter.Events
	inter.ForEachRandEvent(nodes, int(SuperFrameLen)-1, 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)

			inputs[0].SetEvent(e)
			assertar.NoError(
				posets[0].ProcessEvent(e))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			return posets[0].Prepare(e)
		},
	})

	for i := 1; i < len(posets); i++ {
		ee := reorder(ordered)
		for _, e := range ee {
			if e.Epoch != 1 {
				continue
			}
			inputs[i].SetEvent(e)
			posets[i].ProcessEvent(e)
		}
	}

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

				assertar.Equal(*posets[j].checkpoint, *posets[i].checkpoint)
				assertar.Equal(posets[j].superFrame, posets[i].superFrame)

				both := p0.LastBlockN
				if both > p1.LastBlockN {
					both = p1.LastBlockN
				}

				for b := idx.Block(1); b <= both; b++ {
					if !assertar.Equal(
						p0.blocks[b], p1.blocks[b],
						"block %d", b) {
						break
					}
				}

			}
		}
	})
}
