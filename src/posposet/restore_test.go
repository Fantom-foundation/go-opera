package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestRestore(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const posetCount = 3 // 2 last will be restored
	const epochs = idx.SuperFrame(2)

	nodes := inter.GenNodes(5)

	posets := make([]*Poset, 0, posetCount)
	inputs := make([]*EventStore, 0, posetCount)

	makePoset := func(i int) *Store {
		poset, store, input := FakePoset(nodes)
		n := i % len(nodes)
		poset.SetName(nodes[n].String())
		store.SetName(nodes[n].String())
		posets = append(posets, poset.Poset)
		inputs = append(inputs, input)
		return store
	}

	for i := 0; i < posetCount-1; i++ {
		_ = makePoset(i)
	}

	// create events on poset0
	var ordered []*inter.Event
	for epoch := idx.SuperFrame(1); epoch <= epochs; epoch++ {
		r := rand.New(rand.NewSource(int64((epoch))))
		_ = inter.ForEachRandEvent(nodes, int(SuperFrameLen)*3, 3, r, inter.ForEachEvent{
			Process: func(e *inter.Event, name string) {
				inputs[0].SetEvent(e)
				assertar.NoError(posets[0].ProcessEvent(e))

				ordered = append(ordered, e)
			},
			Build: func(e *inter.Event, name string) *inter.Event {
				e.Epoch = epoch
				return posets[0].Prepare(e)
			},
		})
	}

	t.Run("Restore", func(t *testing.T) {

		i := posetCount - 1
		j := posetCount - 2
		store := makePoset(i)

		// use pre-ordered events, call consensus(e) directly, to avoid issues with restoring state of EventBuffer
		for x, e := range ordered {
			if (x < len(ordered)/4) || x%20 == 0 {
				// restore
				restored := New(store, inputs[i])
				n := i % len(nodes)
				restored.SetName("restored_" + nodes[n].String())
				store.SetName("restored_" + nodes[n].String())
				restored.Bootstrap(nil)
				posets[i] = restored
			}
			// push on restore i, and non-restored j
			inputs[i].SetEvent(e)
			assertar.NoError(posets[i].ProcessEvent(e))

			inputs[j].SetEvent(e)
			assertar.NoError(posets[j].ProcessEvent(e))
			// compare state on i/j
			assertar.Equal(*posets[j].checkpoint, *posets[i].checkpoint)
			assertar.Equal(posets[j].superFrame, posets[i].superFrame)
		}
	})
}
