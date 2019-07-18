package posposet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

func TestPoset(t *testing.T) {
	logger.SetTestMode(t)

	const posetCount = 3

	nodes, events := inter.GenEventsByNode(5, 99, 3)

	posets := make([]*Poset, posetCount)
	inputs := make([]*EventStore, posetCount)
	for i := 0; i < posetCount; i++ {
		posets[i], _, inputs[i] = FakePoset(nodes)
		posets[i].SetName(nodes[i].String())
		posets[i].store.SetName(nodes[i].String())
		posets[i].Start()
	}

	t.Run("Multiple start", func(t *testing.T) {
		posets[0].Stop()
		posets[0].Start()
		posets[0].Start()
	})

	t.Run("Push unordered events", func(t *testing.T) {
		// first all events from one node
		for i := 0; i < posetCount; i++ {
			n := i % len(nodes)
			ee := events[nodes[n]]
			for _, e := range ee {
				inputs[i].SetEvent(e)
				posets[i].PushEventSync(e.Hash())
			}
		}
		// second all events from others
		for i := 0; i < posetCount; i++ {
			for n := range nodes {
				if n == i%len(nodes) {
					continue
				}
				ee := events[nodes[n]]
				for _, e := range ee {
					inputs[i].SetEvent(e)
					posets[i].PushEventSync(e.Hash())
				}
			}
		}
	})

	t.Run("Check consensus", func(t *testing.T) {
		assertar := assert.New(t)
		for i := 0; i < len(posets)-1; i++ {
			p0 := posets[i]
			st0 := p0.store.GetCheckpoint()
			t.Logf("Compare poset%d: frame %d, block %d", i, st0.LastDecidedFrameN, st0.LastBlockN)
			for j := i + 1; j < len(posets); j++ {
				p1 := posets[j]
				st1 := p1.store.GetCheckpoint()
				t.Logf("with poset%d: frame %d, block %d", j, st1.LastDecidedFrameN, st1.LastBlockN)

				both := p0.LastBlockN
				if both > p1.LastBlockN {
					both = p1.LastBlockN
				}

				var failAt idx.Block
				for b := idx.Block(1); b <= both; b++ {
					if !assertar.Equal(
						p0.store.GetBlock(b).Events, p1.store.GetBlock(b).Events,
						"block %d", b) {
						failAt = b
						break
					}
				}
				if failAt == 0 {
					continue
				}

				scheme0, err := inter.DAGtoASCIIscheme(p0.EventsTillBlock(failAt))
				if err != nil {
					t.Fatal(err)
				}

				scheme1, err := inter.DAGtoASCIIscheme(p1.EventsTillBlock(failAt))
				if err != nil {
					t.Fatal(err)
				}

				DAGs := utils.TextColumns(scheme0, scheme1)
				t.Log(DAGs)

				return
			}
		}
	})

	t.Run("Multiple stop", func(t *testing.T) {
		posets[0].Stop()
		posets[0].Stop()
	})

	for i := 0; i < posetCount; i++ {
		posets[i].Stop()
	}
}
