package poset

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

// TestPoset 's possibility to get consensus in general on any event order.
func TestPoset(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	const posetCount = 3
	nodes := inter.GenNodes(5)

	posets := make([]*ExtendedPoset, 0, posetCount)
	inputs := make([]*EventStore, 0, posetCount)
	for i := 0; i < posetCount; i++ {
		poset, store, input := FakePoset("", nodes)
		n := i % len(nodes)
		poset.SetName(hash.GetNodeName(nodes[n]))
		store.SetName(hash.GetNodeName(nodes[n]))
		posets = append(posets, poset)
		inputs = append(inputs, input)
	}

	// create events on poset0
	var ordered inter.Events
	inter.ForEachRandEvent(nodes, int(posets[0].dag.MaxEpochBlocks)-1, 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)

			inputs[0].SetEvent(e)
			assertar.NoError(
				posets[0].ProcessEvent(e))
			assertar.NoError(
				flushDb(posets[0], e.Hash()))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			if e.Seq%2 != 0 {
				e.Transactions = append(e.Transactions, &types.Transaction{})
			}
			e.TxHash = types.DeriveSha(e.Transactions)
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
			assertar.NoError(
				posets[i].ProcessEvent(e))
			assertar.NoError(
				flushDb(posets[i], e.Hash()))
		}
	}

	t.Run("Check consensus", func(t *testing.T) {
		compareResults(t, posets)
	})
}

// reorder events, but ancestors are before it's descendants.
func reorder(events inter.Events) inter.Events {
	unordered := make(inter.Events, len(events))
	for i, j := range rand.Perm(len(events)) {
		unordered[j] = events[i]
	}

	reordered := unordered.ByParents()
	return reordered
}

func compareResults(t *testing.T, posets []*ExtendedPoset) {
	assertar := assert.New(t)

	for i := 0; i < len(posets)-1; i++ {
		p0 := posets[i]
		st0 := p0.store.GetCheckpoint()
		ep0 := p0.store.GetEpoch()
		t.Logf("Compare poset%d: Epoch %d, Block %d", i, ep0.EpochN, st0.LastBlockN)
		for j := i + 1; j < len(posets); j++ {
			p1 := posets[j]
			st1 := p1.store.GetCheckpoint()
			t.Logf("with poset%d: Epoch %d, Block %d", j, ep0.EpochN, st1.LastBlockN)

			assertar.Equal(*posets[j].Checkpoint, *posets[i].Checkpoint)
			assertar.Equal(posets[j].EpochState, posets[i].EpochState)

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
}
