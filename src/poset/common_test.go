package poset

import (
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
)

var (
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
)

const (
	enough = 1000000000
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
func FakePoset(namespace string, nodes []common.Address) (*ExtendedPoset, *Store, *EventStore) {
	balances := make(genesis.Accounts, len(nodes))
	for _, addr := range nodes {
		balances[addr] = genesis.Account{Balance: pos.StakeToBalance(1)}
	}

	fs := newFakeFS(namespace)
	store := NewStore(fs.OpenFakeDB(""), fs.OpenFakeDB)

	err := store.ApplyGenesis(&genesis.Genesis{
		Alloc: balances,
		Time:  genesisTestTime,
	}, hash.ZeroEvent, common.Hash{})
	if err != nil {
		panic(err)
	}

	input := NewEventStore(nil)

	poset := New(lachesis.FakeNetDagConfig(), store, input)

	extended := &ExtendedPoset{
		Poset:  poset,
		blocks: map[idx.Block]*inter.Block{},
	}

	extended.Bootstrap(func(block *inter.Block, stateHash common.Hash, members pos.Members) (common.Hash, pos.Members) {
		// track block events
		if extended.blocks[block.Index] != nil {
			extended.Log.Crit("Created block twice")
		}
		extended.blocks[block.Index] = block
		return stateHash, members
	})

	return extended, store, input
}
