package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
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
		buildEvent := func(e *inter.Event) *inter.Event {
			e.Epoch = epoch
			return posets[0].Prepare(e)
		}
		onNewEvent := func(e *inter.Event) {
			inputs[0].SetEvent(e)
			assertar.NoError(posets[0].ProcessEvent(e))

			ordered = append(ordered, e)
		}
		r := rand.New(rand.NewSource(int64((epoch))))
		_ = inter.GenEventsByNode(nodes, int(SuperFrameLen)*3, 3, buildEvent, onNewEvent, r)
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
				restored.Bootstrap()
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

			// check heads
			if e.Seq <= 1 || e.Epoch != posets[i].SuperFrameN {
				continue
			}
			prevEvent := ordered[x-1].Hash()
			assertar.True(posets[i].store.IsHead(e.Hash()), e.Hash().String()) // it's the one the heads, because it's just connected
			for _, p := range e.Parents {
				if p == prevEvent {
					prevEvent = hash.ZeroEvent // not head anymore
				}
				assertar.False(posets[i].store.IsHead(p), p.String()) // not head anymore
			}
			if !prevEvent.IsZero() {
				assertar.True(posets[i].store.IsHead(prevEvent), prevEvent.String())
			}
		}
	})
}
