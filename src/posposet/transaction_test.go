package posposet

import (
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

	// force Epoch commit
	p.nextEpoch(hash.ZeroEvent)

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
