package posposet

import (
	"math/rand"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

var (
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
)

// FakePoset creates empty poset with mem store and equal stakes of nodes in genesis.
func FakePoset(nodes []hash.Peer) (*Poset, *Store, *EventStore) {
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

	return poset, store, input
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
