package poset

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/stretchr/testify/assert"
)

func TestRestore(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const posetCount = 3 // 2 last will be restored
	const epochs = idx.SuperFrame(2)

	nodes := inter.GenNodes(5)

	posets := make([]*ExtendedPoset, 0, posetCount)
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
				restored.Bootstrap(posets[i].applyBlock)
				posets[i].Poset = restored
			}
			// push on restore i, and non-restored j
			inputs[i].SetEvent(e)
			assertar.NoError(posets[i].ProcessEvent(e))

			inputs[j].SetEvent(e)
			assertar.NoError(posets[j].ProcessEvent(e))
			// compare state on i/j
			assertar.Equal(*posets[j].checkpoint, *posets[i].checkpoint)
			assertar.Equal(posets[j].superFrame, posets[i].superFrame)
			// check LastFiWitness and Head() method
			if posets[i].checkpoint.LastBlockN != 0 {
				assertar.Equal(posets[i].checkpoint.LastFiWitness, posets[j].blocks[idx.Block(len(posets[j].blocks))].Head(), "fiWitness must be last event in block")
			}
		}

		// check that blocks are identical
		assertar.Equal(len(posets[i].blocks), len(posets[j].blocks))
		assertar.Equal(len(posets[i].blocks), int(epochs)*int(SuperFrameLen))
		assertar.Equal(len(posets[i].blocks), int(posets[i].LastBlockN))
		for blockI := idx.Block(1); blockI <= idx.Block(len(posets[i].blocks)); blockI++ {
			assertar.NotNil(posets[i].blocks[blockI])
			if t.Failed() {
				return
			}
			assertar.Equal(posets[i].blocks[blockI], posets[j].blocks[blockI])
		}
	})
}
