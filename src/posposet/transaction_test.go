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

	first := true
	buildEvent := func(e *inter.Event) *inter.Event {
		e.Epoch = 1
		e = p.Prepare(e)
		if first && e.Creator == nodes[0] {
			first = false

			e.InternalTransactions = append(e.InternalTransactions,
				&inter.InternalTransaction{
					Nonce:    0,
					Amount:   1,
					Receiver: nodes[1],
				})
		}
		return e
	}
	onNewEvent := func(e *inter.Event) {
		x.SetEvent(e)
		p.PushEventSync(e.Hash())
	}

	p.Start()
	_ = inter.GenEventsByNode(nodes, 20, 3, buildEvent, onNewEvent)

	assert.Equal(t, idx.SuperFrame(0), p.Genesis.Epoch)
	assert.Equal(t, hash.ZeroEvent, p.Genesis.LastFiWitness)
	assert.Equal(t, genesisTestTime, p.Genesis.Time)

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

	assert.Equal(t, idx.SuperFrame(1), p.Genesis.Epoch)
	assert.Equal(t, hash.HexToEventHash("0x6099dac580ff18a7055f5c92c2e0717dd4bf9907565df7a8502d0c3dd513b30c"), p.Genesis.LastFiWitness)
	assert.NotEqual(t, genesisTestTime, p.Genesis.Time)

	assert.Equal(t, inter.Stake(5), p.Members.TotalStake())
	assert.Equal(t, inter.Stake(5), p.NextMembers.TotalStake())

	assert.Equal(t, 4, len(p.Members))
	assert.Equal(t, 4, len(p.NextMembers))

	assert.Equal(t, inter.Stake(0), p.Members[nodes[0]])
	assert.Equal(t, inter.Stake(2), p.Members[nodes[1]])
	assert.Equal(t, inter.Stake(0), p.NextMembers[nodes[0]])
	assert.Equal(t, inter.Stake(2), p.NextMembers[nodes[1]])

	st := s.GetCheckpoint()
	t.Logf("poset: SFrame %d, Block %d", st.SuperFrameN, st.LastBlockN)

	assert.Equal(t,
		inter.Stake(0), p.StakeOf(nodes[0]),
		"balance of %s", nodes[0].String())
	assert.Equal(t,
		inter.Stake(2), p.StakeOf(nodes[1]),
		"balance of %s", nodes[1].String())

	p.Stop()
}
