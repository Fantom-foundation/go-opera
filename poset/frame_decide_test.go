package poset

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestConfirmBlockEvents(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	nodes := inter.GenNodes(5)
	poset, _, input := FakePoset("", nodes)

	var (
		frames []idx.Frame
		blocks []*inter.Block
	)
	applyBlock := poset.applyBlock
	poset.applyBlock = func(block *inter.Block, stateHash common.Hash, validators pos.Validators) (common.Hash, pos.Validators) {
		frames = append(frames, poset.LastDecidedFrame)
		blocks = append(blocks, block)

		return applyBlock(block, stateHash, validators)
	}

	eventCount := int(poset.dag.EpochLen)
	_ = inter.ForEachRandEvent(nodes, eventCount, poset.dag.MaxParents, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			input.SetEvent(e)
			assertar.NoError(
				poset.ProcessEvent(e))
			assertar.NoError(
				flushDb(poset, e.Hash()))

		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = idx.Epoch(1)
			return poset.Prepare(e)
		},
	})

	// unconfirm all events
	it := poset.store.table.ConfirmedEvent.NewIterator()
	batch := poset.store.table.ConfirmedEvent.NewBatch()
	for it.Next() {
		assertar.NoError(batch.Delete(it.Key()))
	}
	assertar.NoError(batch.Write())
	it.Release()

	for i, block := range blocks {
		frame := frames[i]
		atropos := poset.LastAtropos
		if (i + 1) < len(blocks) {
			atropos = blocks[i+1].PrevHash
		}

		// call confirmBlock again
		gotBlock, _ := poset.confirmBlock(frame, atropos, nil)

		if !assertar.Equal(block.Events, gotBlock.Events) {
			break
		}
	}
}
