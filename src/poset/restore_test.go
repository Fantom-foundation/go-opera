package poset

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/fallible"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestRestore(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const (
		COUNT     = 3 // two poset instances
		GENERATOR = 0 // event generator
		EXPECTED  = 1 // first as etalon
		RESTORED  = 2 // second with db failures
	)

	nodes := inter.GenNodes(5)
	posets := make([]*ExtendedPoset, 0, COUNT)
	inputs := make([]*EventStore, 0, COUNT)
	namespaces := make([]string, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		namespace := uniqNamespace()
		poset, _, input := FakePoset(namespace, nodes)
		posets = append(posets, poset)
		inputs = append(inputs, input)
		namespaces = append(namespaces, namespace)
	}

	posets[GENERATOR].
		SetName("generator")
	posets[GENERATOR].store.
		SetName("generator")

	const epochs = 2
	var epochLen = int(posets[GENERATOR].dag.EpochLen)

	// create events
	var ordered []*inter.Event
	for epoch := idx.Epoch(1); epoch <= idx.Epoch(epochs); epoch++ {
		r := rand.New(rand.NewSource(int64((epoch))))
		parentCount := epochLen * COUNT
		_ = inter.ForEachRandEvent(nodes, parentCount, COUNT, r, inter.ForEachEvent{
			Process: func(e *inter.Event, name string) {
				inputs[GENERATOR].SetEvent(e)
				assertar.NoError(posets[GENERATOR].ProcessEvent(e))

				ordered = append(ordered, e)
			},
			Build: func(e *inter.Event, name string) *inter.Event {
				e.Epoch = epoch
				return posets[GENERATOR].Prepare(e)
			},
		})
	}

	posets[EXPECTED].
		SetName("expected")
	posets[EXPECTED].store.
		SetName("expected")

	posets[RESTORED].
		SetName("restored-0")
	posets[RESTORED].store.
		SetName("restored-0")

	// use pre-ordered events, call consensus(e) directly, to avoid issues with restoring state of EventBuffer
	x := 0
	for n, e := range ordered {
		if n%20 == 0 {
			log.Info("restart poset")
			prev := posets[RESTORED]
			x++
			fs := newFakeFS(namespaces[RESTORED])
			store := NewStore(fs.OpenFakeDB(""), fs.OpenFakeDB)
			store.SetName(fmt.Sprintf("restored-%d", x))

			restored := New(prev.dag, store, prev.input)
			restored.SetName(fmt.Sprintf("restored-%d", x))
			restored.Bootstrap(prev.applyBlock)

			posets[RESTORED].Poset = restored
		}

		inputs[EXPECTED].SetEvent(e)
		assertar.NoError(posets[EXPECTED].ProcessEvent(e))

		inputs[RESTORED].SetEvent(e)
		assertar.NoError(posets[RESTORED].ProcessEvent(e))

		// compare states
		assertar.Equal(
			*posets[EXPECTED].checkpoint, *posets[RESTORED].checkpoint)
		assertar.Equal(
			posets[EXPECTED].epochState.PrevEpoch.Hash(), posets[RESTORED].epochState.PrevEpoch.Hash())
		assertar.Equal(
			posets[EXPECTED].epochState.Members, posets[RESTORED].epochState.Members)
		assertar.Equal(
			posets[EXPECTED].epochState.EpochN, posets[RESTORED].epochState.EpochN)
		// check LastAtropos and Head() method
		if posets[EXPECTED].checkpoint.LastBlockN != 0 {
			assertar.Equal(
				posets[RESTORED].checkpoint.LastAtropos,
				posets[EXPECTED].blocks[idx.Block(len(posets[EXPECTED].blocks))].Hash(),
				"atropos must be last event in block")
		}
	}

	// check that blocks are identical
	assertar.Equal(len(posets[EXPECTED].blocks), len(posets[RESTORED].blocks))
	assertar.Equal(len(posets[EXPECTED].blocks), epochLen*epochs)
	assertar.Equal(len(posets[EXPECTED].blocks), int(posets[RESTORED].LastBlockN))
	for i := idx.Block(1); i <= idx.Block(len(posets[RESTORED].blocks)); i++ {
		assertar.NotNil(posets[RESTORED].blocks[i])
		if t.Failed() {
			return
		}
		assertar.Equal(posets[EXPECTED].blocks[i], posets[RESTORED].blocks[i])
	}
}

func TestDbFailure(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const (
		COUNT    = 2 // two poset instances
		EXPECTED = 0 // first as etalon
		RESTORED = 1 // second with db failures
	)
	nodes := inter.GenNodes(5)

	posets := make([]*ExtendedPoset, 0, COUNT)
	inputs := make([]*EventStore, 0, COUNT)
	namespaces := make([]string, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		namespace := uniqNamespace()
		poset, _, input := FakePoset(namespace, nodes)
		posets = append(posets, poset)
		inputs = append(inputs, input)
		namespaces = append(namespaces, namespace)
	}

	posets[EXPECTED].
		SetName("expected")
	posets[EXPECTED].store.
		SetName("expected")

	// create events on etalon poset
	var ordered inter.Events
	inter.ForEachRandEvent(nodes, int(posets[EXPECTED].dag.EpochLen)-1, 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)

			inputs[EXPECTED].SetEvent(e)
			assertar.NoError(
				posets[EXPECTED].ProcessEvent(e))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			return posets[EXPECTED].Prepare(e)
		},
	})

	posets[RESTORED].
		SetName("restored-0")
	posets[RESTORED].store.
		SetName("restored-0")

	// db writes limit
	db := posets[RESTORED].store.UnderlyingDB().(*fallible.Fallible)
	db.SetWriteCount(100) // TODO: test all stages fault

	x := 0
	process := func(e *inter.Event) (ok bool) {
		ok = true
		defer func() {
			// catch a panic
			if r := recover(); r == nil {
				return
			}
			ok = false

			db.SetWriteCount(enough)

			t.Log("restart poset after db failure")
			prev := posets[RESTORED]
			x++
			fs := newFakeFS(namespaces[RESTORED])
			store := NewStore(fs.OpenFakeDB(""), fs.OpenFakeDB)
			store.SetName(fmt.Sprintf("restored-%d", x))

			restored := New(prev.dag, store, prev.input)
			restored.SetName(fmt.Sprintf("restored-%d", x))
			restored.Bootstrap(prev.applyBlock)

			posets[RESTORED].Poset = restored
		}()

		inputs[RESTORED].SetEvent(e)
		assertar.NoError(
			posets[RESTORED].ProcessEvent(e))

		return
	}

	for len(ordered) > 0 {
		e := ordered[0]
		if e.Epoch != 1 {
			continue
		}

		if process(e) {
			ordered = ordered[1:]
		}
	}

	t.Run("Check consensus", func(t *testing.T) {
		compareResults(t, posets)
	})

}
