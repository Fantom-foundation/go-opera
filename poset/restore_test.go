package poset

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/fallible"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestRestore(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const (
		COUNT     = 3 // two poset instances
		GENERATOR = 0 // event generator
		EXPECTED  = 1 // first as etalon
		RESTORED  = 2 // second with restoring
	)

	nodes := inter.GenNodes(5)
	posets := make([]*ExtendedPoset, 0, COUNT)
	inputs := make([]*EventStore, 0, COUNT)
	namespaces := make([]string, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		namespace := fmt.Sprintf("poset.TestRestore-%d-%d", i, rand.Int())
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
	var epochLen = 30

	// seal epoch on decided frame == epochLen
	for _, poset := range posets {
		applyBlock := poset.callback.ApplyBlock
		poset.callback.ApplyBlock = func(block *inter.Block, decidedFrame idx.Frame, cheaters inter.Cheaters) (common.Hash, bool) {
			h, _ := applyBlock(block, decidedFrame, cheaters)
			return h, decidedFrame == idx.Frame(epochLen)
		}
	}

	// create events
	var ordered []*inter.Event
	for epoch := idx.Epoch(1); epoch <= idx.Epoch(epochs); epoch++ {
		stability := rand.New(rand.NewSource(int64(epoch)))
		_ = inter.ForEachRandEvent(nodes, epochLen*4, COUNT, stability, inter.ForEachEvent{
			Process: func(e *inter.Event, name string) {
				inputs[GENERATOR].SetEvent(e)
				assertar.NoError(
					posets[GENERATOR].ProcessEvent(e))
				assertar.NoError(
					flushDb(posets[GENERATOR], e.Hash()))

				ordered = append(ordered, e)
			},
			Build: func(e *inter.Event, name string) *inter.Event {
				e.Epoch = epoch
				if e.Seq%2 != 0 {
					e.Transactions = append(e.Transactions, &types.Transaction{})
				}
				e.TxHash = types.DeriveSha(e.Transactions)
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
			log.Info("Restart poset")
			prev := posets[RESTORED]
			x++

			mems := memorydb.NewProducer(namespaces[RESTORED])
			dbs := flushable.NewSyncedPool(mems)
			store := NewStore(dbs, LiteStoreConfig())
			store.SetName(fmt.Sprintf("restored-%d", x))

			restored := New(prev.dag, store, prev.input)
			restored.SetName(fmt.Sprintf("restored-%d", x))

			restored.Bootstrap(prev.callback)

			posets[RESTORED].Poset = restored
		}

		inputs[EXPECTED].SetEvent(e)
		assertar.NoError(
			posets[EXPECTED].ProcessEvent(e))
		assertar.NoError(
			flushDb(posets[EXPECTED], e.Hash()))

		inputs[RESTORED].SetEvent(e)
		assertar.NoError(
			posets[RESTORED].ProcessEvent(e))
		assertar.NoError(
			flushDb(posets[RESTORED], e.Hash()))

		compareStates(assertar, posets[EXPECTED], posets[RESTORED])
		if t.Failed() {
			return
		}
	}

	if !assertar.Equal(epochLen*epochs, len(posets[EXPECTED].blocks)) {
		return
	}
	compareBlocks(assertar, posets[EXPECTED], posets[RESTORED])
}

func TestDbFailure(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const (
		COUNT     = 3 // two poset instances
		GENERATOR = 0 // event generator
		EXPECTED  = 1 // first as etalon
		RESTORED  = 2 // second with db failures
	)
	nodes := inter.GenNodes(5)

	var realDb *fallible.Fallible
	setRealDb := func(db kvdb.KeyValueStore) kvdb.KeyValueStore {
		if realDb != nil {
			return db
		}
		fdb := fallible.Wrap(db)
		fdb.SetWriteCount(enough)
		realDb = fdb
		return fdb
	}

	posets := make([]*ExtendedPoset, 0, COUNT)
	inputs := make([]*EventStore, 0, COUNT)
	namespaces := make([]string, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		namespace := fmt.Sprintf("poset.TestDbFailure-%d-%d", i, rand.Int())
		var mods []memorydb.Mod
		if i == RESTORED {
			mods = []memorydb.Mod{setRealDb}
		}
		poset, _, input := FakePoset(namespace, nodes, mods...)
		posets = append(posets, poset)
		inputs = append(inputs, input)
		namespaces = append(namespaces, namespace)
	}

	posets[GENERATOR].
		SetName("generator")
	posets[GENERATOR].store.
		SetName("generator")

	epochLen := int(posets[GENERATOR].dag.EpochLen)

	stability := rand.New(rand.NewSource(1))
	// create events on etalon poset
	var ordered inter.Events
	inter.ForEachRandEvent(nodes, epochLen-1, COUNT, stability, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)

			inputs[GENERATOR].SetEvent(e)
			assertar.NoError(
				posets[GENERATOR].ProcessEvent(e))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			if e.Seq%2 != 0 {
				e.Transactions = append(e.Transactions, &types.Transaction{})
			}
			e.TxHash = types.DeriveSha(e.Transactions)
			return posets[GENERATOR].Prepare(e)
		},
	})

	posets[EXPECTED].
		SetName("expected")
	posets[EXPECTED].store.
		SetName("expected")

	posets[RESTORED].
		SetName("restored-0")
	posets[RESTORED].store.
		SetName("restored-0")

	x := 0
	process := func(e *inter.Event) (ok bool) {
		ok = true
		defer func() {
			// catch a panic
			if r := recover(); r == nil {
				return
			}
			ok = false

			log.Info("Restart poset after db failure")
			prev := posets[RESTORED]
			x++

			realDb.SetWriteCount(100)
			mems := memorydb.NewProducer(namespaces[RESTORED])
			dbs := flushable.NewSyncedPool(mems)
			store := NewStore(dbs, LiteStoreConfig())
			store.SetName(fmt.Sprintf("restored-%d", x))

			restored := New(prev.dag, store, prev.input)
			restored.SetName(fmt.Sprintf("restored-%d", x))
			restored.Bootstrap(prev.callback)

			posets[RESTORED].Poset = restored
		}()

		inputs[RESTORED].SetEvent(e)
		assertar.NoError(
			posets[RESTORED].ProcessEvent(e))
		assertar.NoError(
			flushDb(posets[RESTORED], e.Hash()))

		inputs[EXPECTED].SetEvent(e)
		assertar.NoError(
			posets[EXPECTED].ProcessEvent(e))
		assertar.NoError(
			flushDb(posets[EXPECTED], e.Hash()))

		return
	}

	for len(ordered) > 0 {
		e := ordered[0]
		if e.Epoch != 1 {
			panic("sanity check")
		}

		if !process(e) {
			continue
		}

		ordered = ordered[1:]

		compareStates(assertar, posets[EXPECTED], posets[RESTORED])
		if t.Failed() {
			return
		}
	}

	compareBlocks(assertar, posets[EXPECTED], posets[RESTORED])
}

func compareStates(assertar *assert.Assertions, expected, restored *ExtendedPoset) {
	assertar.Equal(
		*expected.Checkpoint, *restored.Checkpoint)
	assertar.Equal(
		expected.EpochState.PrevEpoch.Hash(), restored.EpochState.PrevEpoch.Hash())
	assertar.Equal(
		expected.EpochState.Validators, restored.EpochState.Validators)
	assertar.Equal(
		expected.EpochState.EpochN, restored.EpochState.EpochN)
	// check LastAtropos and Head() method
	if expected.Checkpoint.LastBlockN != 0 {
		assertar.Equal(
			expected.blocks[idx.Block(len(expected.blocks))].Hash(),
			restored.Checkpoint.LastAtropos,
			"atropos must be last event in block")
	}
}

func compareBlocks(assertar *assert.Assertions, expected, restored *ExtendedPoset) {
	assertar.Equal(len(expected.blocks), len(restored.blocks))
	assertar.Equal(len(expected.blocks), int(restored.LastBlockN))
	for i := idx.Block(1); i <= idx.Block(len(restored.blocks)); i++ {
		if !assertar.NotNil(restored.blocks[i]) ||
			!assertar.Equal(expected.blocks[i], restored.blocks[i]) {
			return
		}

	}
}
