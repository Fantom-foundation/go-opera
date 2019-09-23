package poset

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/fallible"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestRestore(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const posetCount = 3 // 2 last will be restored
	const epochs = idx.Epoch(2)

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
	for epoch := idx.Epoch(1); epoch <= epochs; epoch++ {
		r := rand.New(rand.NewSource(int64((epoch))))
		_ = inter.ForEachRandEvent(nodes, int(posets[0].dag.EpochLen)*3, 3, r, inter.ForEachEvent{
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
				restored := New(posets[0].dag, store, inputs[i])
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
			assertar.Equal(posets[j].epochState.PrevEpoch.Hash(), posets[i].epochState.PrevEpoch.Hash())
			assertar.Equal(posets[j].epochState.Members, posets[i].epochState.Members)
			assertar.Equal(posets[j].epochState.EpochN, posets[i].epochState.EpochN)
			// check LastAtropos and Head() method
			if posets[i].checkpoint.LastBlockN != 0 {
				assertar.Equal(posets[i].checkpoint.LastAtropos, posets[j].blocks[idx.Block(len(posets[j].blocks))].Hash(), "atropos must be last event in block")
			}
		}

		// check that blocks are identical
		assertar.Equal(len(posets[i].blocks), len(posets[j].blocks))
		assertar.Equal(len(posets[i].blocks), int(epochs)*int(posets[0].dag.EpochLen))
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

func TestDbFailure(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const (
		CNT = 2 // two poset instances
		EXP = 0 // first as etalon
		GOT = 1 // second with db failures
	)
	nodes := inter.GenNodes(5)

	posets := make([]*ExtendedPoset, 0, CNT)
	inputs := make([]*EventStore, 0, CNT)

	makePoset := func(i int) *Store {
		poset, store, input := FakePoset(nodes)
		n := i % len(nodes)
		poset.SetName(nodes[n].String())
		store.SetName(nodes[n].String())
		posets = append(posets, poset)
		inputs = append(inputs, input)
		return store
	}
	for i := 0; i < CNT; i++ {
		makePoset(i)
	}

	// create events on etalon poset
	var ordered inter.Events
	inter.ForEachRandEvent(nodes, int(posets[EXP].dag.EpochLen)-1, 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)

			inputs[EXP].SetEvent(e)
			assertar.NoError(
				posets[EXP].ProcessEvent(e))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			return posets[EXP].Prepare(e)
		},
	})

	// db writes limit
	db := posets[GOT].store.UnderlyingDB().(*fallible.Fallible)
	db.SetWriteCount(10000) // TODO: test all stages

	for len(ordered) > 0 {
		e := ordered[0]
		if e.Epoch != 1 {
			continue
		}

		func() {
			defer func() {
				// recreate poset on panic
				if r := recover(); r != nil {
					t.Log("restart poset after db failure")
					prev := posets[GOT]

					db.SetWriteCount(enough)
					store := NewStore(db, func(name string) kvdb.KeyValueStore {
						return memorydb.New()
					})

					restored := New(prev.dag, store, prev.input)
					restored.SetName("restored")
					store.SetName("restored")
					restored.Bootstrap(prev.applyBlock)

					posets[GOT] = &ExtendedPoset{
						Poset:  restored,
						blocks: map[idx.Block]*inter.Block{},
					}
				}
			}()

			inputs[GOT].SetEvent(e)
			posets[GOT].ProcessEvent(e)
		}()

		ordered = ordered[1:]
	}

	t.Run("Check consensus", func(t *testing.T) {
		compareResults(t, posets)
	})

}
