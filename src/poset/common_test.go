package posposet

import (
	"math/rand"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

var (
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
)

// ExtendedPoset extends Poset for tests.
type ExtendedPoset struct {
	*Poset

	blocks map[idx.Block]*inter.Block
}

func (p *ExtendedPoset) EventsTillBlock(until idx.Block) hash.Events {
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
func FakePoset(nodes []hash.Peer) (*ExtendedPoset, *Store, *EventStore) {
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

	extended := &ExtendedPoset{
		Poset: poset,
	}
	// track block events
	extended.applyBlock = func(block *inter.Block, stateHash hash.Hash, members pos.Members) (hash.Hash, pos.Members) {
		if extended.blocks == nil {
			extended.blocks = map[idx.Block]*inter.Block{}
		}
		if extended.blocks[block.Index] != nil {
			extended.Fatal("created block twice")
		}
		extended.blocks[block.Index] = block
		return stateHash, members
	}

	return extended, store, input
}

func reorder(events inter.Events) inter.Events {
	count := len(events)
	unordered := make(inter.Events, count)
	pos := rand.Perm(count)
	for i := 0; i < count; i++ {
		unordered[pos[i]] = events[i]
	}

	reordered := unordered.ByParents()
	return reordered
}
