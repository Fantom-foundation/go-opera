package posposet

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestPosetTxn(t *testing.T) {
	first := true
	transfer := func(e *inter.Event, nodes []hash.Peer) {
		if !first {
			return
		}
		if e.Creator != nodes[0] {
			return
		}
		first = false

		e.InternalTransactions = append(e.InternalTransactions,
			&inter.InternalTransaction{
				Index:    0,
				Amount:   1,
				Receiver: nodes[1],
			})
	}

	nodes, events := inter.GenEventsByNode(5, 20, 3, transfer)

	p, s, x := FakePoset(nodes)
	assert.Equal(t,
		uint64(1), p.StakeOf(nodes[0]),
		"balance of %s", nodes[0].String())
	assert.Equal(t,
		uint64(1), p.StakeOf(nodes[1]),
		"balance of %s", nodes[1].String())

	p.Start()
	for _, n := range nodes {
		for _, e := range events[n] {
			x.SetEvent(e)
			p.PushEventSync(e.Hash())
		}
	}

	st := s.GetState()
	t.Logf("poset: frame %d, block %d", atomic.LoadUint64(&st.LastFinishedFrameN), st.LastBlockN)

	assert.Equal(t,
		uint64(0), p.StakeOf(nodes[0]),
		"balance of %s", nodes[0].String())
	assert.Equal(t,
		uint64(2), p.StakeOf(nodes[1]),
		"balance of %s", nodes[1].String())
	p.Stop()
}
