package poset

import (
	"bytes"
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/logger"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

func TestConfirmBlockEvents(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	fmt.Println("DBG1")

	nodes := inter.GenNodes(5)
	poset, _, input := FakePoset("", nodes)

	fmt.Println("DBG2")

	var (
		frames []idx.Frame
		blocks []*inter.Block
	)
	applyBlock := poset.applyBlock
	fmt.Println("DBG3")
	poset.applyBlock = func(block *inter.Block, stateHash common.Hash, validators pos.Validators) (common.Hash, pos.Validators) {
		frames = append(frames, poset.LastDecidedFrame)
		blocks = append(blocks, block)

		return applyBlock(block, stateHash, validators)
	}

	fmt.Println("DBG4")
	_ = inter.ForEachRandEvent(nodes, int(poset.dag.EpochLen), poset.dag.MaxParents, nil, inter.ForEachEvent{
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

	fmt.Println("DBG5")

	for i, block := range blocks {
		frame := frames[i]
		atropos := poset.LastAtropos
		if (i + 1) < len(blocks) {
			atropos = blocks[i+1].PrevHash
		}

		// call confirmBlockEvents again
		unordered, _ := poset.confirmBlockEvents(frame, atropos)

		sort.Slice(unordered, func(i, j int) bool {
			a, b := unordered[i], unordered[j]

			if a.Lamport != b.Lamport {
				return a.Lamport < b.Lamport
			}

			return bytes.Compare(a.Hash().Bytes(), b.Hash().Bytes()) < 0
		})
		ordered := unordered

		got := make(hash.Events, len(ordered))
		for i, e := range ordered {
			got[i] = e.Hash()
		}

		// NOTE: it means confirmBlockEvents() return events once
		expect := hash.Events{}
		// TODO: `expect := block.Events` if confirmBlockEvents() idempotent

		if !assertar.Equal(expect, got, block) {
			break
		}
	}
}
