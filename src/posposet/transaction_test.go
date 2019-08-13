package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestPosetTxn(t *testing.T) {
	logger.SetTestMode(t)

	nodes := inter.GenNodes(5)

	p, s, x := FakePoset(nodes)
	assert.Equal(t,
		inter.Stake(1), p.StakeOf(nodes[0]),
		"balance of %s", nodes[0].String())
	assert.Equal(t,
		inter.Stake(1), p.StakeOf(nodes[1]),
		"balance of %s", nodes[1].String())

	buildEvent := func(e *inter.Event) *inter.Event {
		e.Epoch = 1
		if e.Seq == 1 && e.Creator == nodes[0] {
			e.InternalTransactions = append(e.InternalTransactions,
				&inter.InternalTransaction{
					Nonce:    0,
					Amount:   1,
					Receiver: nodes[1],
				})
		}
		e = p.Prepare(e)
		return e
	}
	onNewEvent := func(e *inter.Event) {
		x.SetEvent(e)
		assert.NoError(t, p.ProcessEvent(e))
	}

	_ = inter.GenEventsByNode(nodes, int(SuperFrameLen-1), 3, buildEvent, onNewEvent, nil)

	assert.Equal(t, p.PrevEpoch.Hash(), s.GetGenesis().PrevEpoch.Hash())

	assert.Equal(t, idx.SuperFrame(0), p.PrevEpoch.Epoch)
	assert.Equal(t, hash.ZeroEvent, p.PrevEpoch.LastFiWitness)
	assert.Equal(t, genesisTestTime, p.PrevEpoch.Time)

	assert.Equal(t, inter.Stake(5), p.Members.TotalStake())
	assert.Equal(t, inter.Stake(5), p.NextMembers.TotalStake())

	assert.Equal(t, 5, len(p.Members))
	assert.Equal(t, 4, len(p.NextMembers))

	assert.Equal(t, inter.Stake(1), p.Members[nodes[0]])
	assert.Equal(t, inter.Stake(1), p.Members[nodes[1]])
	assert.Equal(t, inter.Stake(0), p.NextMembers[nodes[0]])
	assert.Equal(t, inter.Stake(2), p.NextMembers[nodes[1]])

	// force Epoch commit
	p.nextEpoch(hash.HexToEventHash("0x6099dac580ff18a7055f5c92c2e0717dd4bf9907565df7a8502d0c3dd513b30c"))

	assert.Equal(t, idx.SuperFrame(1), p.PrevEpoch.Epoch)
	assert.Equal(t, hash.HexToEventHash("0x6099dac580ff18a7055f5c92c2e0717dd4bf9907565df7a8502d0c3dd513b30c"), p.PrevEpoch.LastFiWitness)
	assert.NotEqual(t, genesisTestTime, p.PrevEpoch.Time)

	assert.Equal(t, inter.Stake(5), p.Members.TotalStake())
	assert.Equal(t, inter.Stake(5), p.NextMembers.TotalStake())

	assert.Equal(t, 4, len(p.Members))
	assert.Equal(t, 4, len(p.NextMembers))

	assert.Equal(t, inter.Stake(0), p.Members[nodes[0]])
	assert.Equal(t, inter.Stake(2), p.Members[nodes[1]])
	assert.Equal(t, inter.Stake(0), p.NextMembers[nodes[0]])
	assert.Equal(t, inter.Stake(2), p.NextMembers[nodes[1]])

	st := s.GetCheckpoint()
	ep := s.GetSuperFrame()
	t.Logf("poset: SFrame %d, Block %d", ep.SuperFrameN, st.LastBlockN)

	assert.Equal(t,
		inter.Stake(0), p.StakeOf(nodes[0]),
		"balance of %s", nodes[0].String())
	assert.Equal(t,
		inter.Stake(2), p.StakeOf(nodes[1]),
		"balance of %s", nodes[1].String())
}
