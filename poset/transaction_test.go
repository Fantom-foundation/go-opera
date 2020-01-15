package poset

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestPosetTxn(t *testing.T) {
	logger.SetTestMode(t)

	var x = pos.Stake(1)

	nodes := inter.GenNodes(5)

	p, s, in := FakePoset("", nodes)
	assert.Equal(t,
		1*x, p.EpochState.Validators.Get(nodes[0]),
		"balance of %s", hash.GetNodeName(nodes[0]))
	assert.Equal(t,
		1*x, p.EpochState.Validators.Get(nodes[1]),
		"balance of %s", hash.GetNodeName(nodes[1]))

	p.callback.SelectValidatorsGroup = func(oldEpoch, newEpoch idx.Epoch) *pos.Validators {
		if oldEpoch == 1 {
			validators := p.Validators.Builder()
			// move stake from node0 to node1
			validators.Set(nodes[0], 0*x)
			validators.Set(nodes[1], 2*x)
			return validators.Build()
		}
		return p.Validators
	}

	_ = inter.ForEachRandEvent(nodes, int(p.dag.MaxEpochBlocks-1), 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			in.SetEvent(e)
			assert.NoError(t,
				p.ProcessEvent(e))
			assert.NoError(t,
				flushDb(p, e.Hash()))
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			e.TxHash = types.DeriveSha(e.Transactions)
			e = p.Prepare(e)
			return e
		},
	})

	assert.Equal(t, p.PrevEpoch.Hash(), s.GetGenesis().PrevEpoch.Hash())

	assert.Equal(t, idx.Epoch(0), p.PrevEpoch.Epoch)
	assert.Equal(t, genesisTime, p.PrevEpoch.Time)

	assert.Equal(t, 5*x, p.Validators.TotalStake())

	assert.Equal(t, 5, p.Validators.Len())

	assert.Equal(t, 1*x, p.Validators.Get(nodes[0]))
	assert.Equal(t, 1*x, p.Validators.Get(nodes[1]))
	// force Epoch commit
	p.sealEpoch()

	assert.Equal(t, idx.Epoch(1), p.PrevEpoch.Epoch)
	assert.NotEqual(t, genesisTime, p.PrevEpoch.Time)

	assert.Equal(t, 5*x, p.Validators.TotalStake())

	assert.Equal(t, 4, p.Validators.Len())

	assert.Equal(t, 0*x, p.Validators.Get(nodes[0]))
	assert.Equal(t, 2*x, p.Validators.Get(nodes[1]))

	st := s.GetCheckpoint()
	ep := s.GetEpoch()
	t.Logf("poset: Epoch %d, Block %d", ep.EpochN, st.LastBlockN)
}
