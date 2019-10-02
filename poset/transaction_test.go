package poset

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestPosetTxn(t *testing.T) {
	logger.SetTestMode(t)

	nodes := inter.GenNodes(5)

	p, s, x := FakePoset(uniqNamespace(), nodes)
	assert.Equal(t,
		pos.Stake(1), p.epochState.Validators[nodes[0]],
		"balance of %s", nodes[0].String())
	assert.Equal(t,
		pos.Stake(1), p.epochState.Validators[nodes[1]],
		"balance of %s", nodes[1].String())

	p.applyBlock = func(block *inter.Block, stateHash common.Hash, validators pos.Validators) (common.Hash, pos.Validators) {
		if block.Index == 1 {
			// move stake from node0 to node1
			validators.Set(nodes[0], 0)
			validators.Set(nodes[1], 2)
		}
		return stateHash, validators
	}

	_ = inter.ForEachRandEvent(nodes, int(p.dag.EpochLen-1), 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			x.SetEvent(e)
			assert.NoError(t, p.ProcessEvent(e))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			e = p.Prepare(e)
			return e
		},
	})

	assert.Equal(t, p.PrevEpoch.Hash(), s.GetGenesis().PrevEpoch.Hash())

	assert.Equal(t, idx.Epoch(0), p.PrevEpoch.Epoch)
	assert.Equal(t, genesisTestTime, p.PrevEpoch.Time)

	assert.Equal(t, pos.Stake(5), p.Validators.TotalStake())
	assert.Equal(t, pos.Stake(5), p.NextValidators.TotalStake())

	assert.Equal(t, 5, len(p.Validators))
	assert.Equal(t, 4, len(p.NextValidators))

	assert.Equal(t, pos.Stake(1), p.Validators[nodes[0]])
	assert.Equal(t, pos.Stake(1), p.Validators[nodes[1]])
	assert.Equal(t, pos.Stake(0), p.NextValidators[nodes[0]])
	assert.Equal(t, pos.Stake(2), p.NextValidators[nodes[1]])

	// force Epoch commit
	p.onNewEpoch(hash.HexToEventHash("0x6099dac580ff18a7055f5c92c2e0717dd4bf9907565df7a8502d0c3dd513b30c"), nil)

	assert.Equal(t, idx.Epoch(1), p.PrevEpoch.Epoch)
	assert.Equal(t, hash.HexToEventHash("0x6099dac580ff18a7055f5c92c2e0717dd4bf9907565df7a8502d0c3dd513b30c"), p.PrevEpoch.LastAtropos)
	assert.NotEqual(t, genesisTestTime, p.PrevEpoch.Time)

	assert.Equal(t, pos.Stake(5), p.Validators.TotalStake())
	assert.Equal(t, pos.Stake(5), p.NextValidators.TotalStake())

	assert.Equal(t, 4, len(p.Validators))
	assert.Equal(t, 4, len(p.NextValidators))

	assert.Equal(t, pos.Stake(0), p.Validators[nodes[0]])
	assert.Equal(t, pos.Stake(2), p.Validators[nodes[1]])
	assert.Equal(t, pos.Stake(0), p.NextValidators[nodes[0]])
	assert.Equal(t, pos.Stake(2), p.NextValidators[nodes[1]])

	st := s.GetCheckpoint()
	ep := s.GetEpoch()
	t.Logf("poset: Epoch %d, Block %d", ep.EpochN, st.LastBlockN)
}
