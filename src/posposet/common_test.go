package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

var (
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
)

type BufferedPoset struct {
	*Poset

	blocks map[idx.Block]*inter.Block

	bufferPush ordering.PushEventFn
}

func (p *BufferedPoset) PushToBuffer(e *inter.Event) {
	p.bufferPush(e, "")
}

func (p *BufferedPoset) EventsTillBlock(until idx.Block) hash.Events {
	res := make(hash.Events, 0)
	for i := idx.Block(1); i <= until; i++ {
		if p.blocks[i] == nil {
			break
		}
		res = append(res, p.blocks[i].Events...)
	}
	return res
}

// FakePoset creates empty poset with mem store and equal stakes of nodes in genesis.
// Input event order doesn't matter.
func FakePoset(nodes []hash.Peer) (*BufferedPoset, *Store, *EventStore) {
	balances := make(map[hash.Peer]pos.Stake, len(nodes))
	for _, addr := range nodes {
		balances[addr] = pos.Stake(1)
	}

	store := NewMemStore()
	err := store.ApplyGenesis(&genesis.Config{
		Balances: balances,
		Time:     genesisTestTime,
	})
	if err != nil {
		panic(err)
	}

	input := NewEventStore(nil)

	poset := New(store, input)
	poset.Bootstrap(nil)

	buffered := &BufferedPoset{
		Poset:      poset,
		bufferPush: MakeOrderedInput(poset),
	}
	// track block events
	buffered.applyBlock = func(block *inter.Block, stateHash hash.Hash, members pos.Members) (hash.Hash, pos.Members) {
		if buffered.blocks == nil {
			buffered.blocks = map[idx.Block]*inter.Block{}
		}
		if buffered.blocks[block.Index] != nil {
			buffered.Fatal("created block twice")
		}
		buffered.blocks[block.Index] = block
		return stateHash, members
	}

	return buffered, store, input
}

// MakeOrderedInput wraps Poset.onNewEvent with ordering.EventBuffer.
// For tests only.
func MakeOrderedInput(p *Poset) ordering.PushEventFn {
	processed := make(hash.EventsSet) // NOTE: mem leak, so for tests only.

	orderThenConsensus, _ := ordering.EventBuffer(ordering.Callback{

		Process: func(event *inter.Event) error {
			processed.Add(event.Hash())
			return p.ProcessEvent(event)
		},

		Drop: func(e *inter.Event, peer string, err error) {
			logger.Get().Warn(err.Error() + ", so rejected")
		},

		Exists: func(h hash.Event) *inter.Event {
			if _, ok := processed[h]; ok {
				return p.input.GetEvent(h)
			}
			return nil
		},
	})
	return orderThenConsensus
}

// ASCIIschemeToDAG wrap inter.ASCIIschemeForEach() to prepare events properly.
func ASCIIschemeToDAG(
	scheme string,
) (
	nodes []hash.Peer,
	events map[hash.Peer][]*inter.Event,
	names map[string]*inter.Event,
) {
	// get nodes only
	nodes, _, _ = inter.ASCIIschemeToDAG(scheme)
	// init poset
	p, _, input := FakePoset(nodes)

	// process events
	nodes, events, names = inter.ASCIIschemeForEach(scheme, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			input.SetEvent(e)
			err := p.ProcessEvent(e)
			if err != nil {
				p.Fatal(err)
			}
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = p.CurrentSuperFrameN()
			e = p.Prepare(e)

			return e
		},
	})

	return
}
