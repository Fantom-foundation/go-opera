package poset

import (
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
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
func FakePoset(namespace string, nodes []common.Address, mods ...memorydb.Mod) (*ExtendedPoset, *Store, *EventStore) {
	balances := make(genesis.Accounts, len(nodes))
	for _, addr := range nodes {
		balances[addr] = genesis.Account{Balance: pos.StakeToBalance(1)}
	}

	mems := memorydb.NewProducer(namespace, mods...)
	dbs := flushable.NewSyncedPool(mems)
	store := NewStore(dbs)

	atropos := hash.ZeroEvent
	err := store.ApplyGenesis(&genesis.Genesis{
		Alloc: balances,
		Time:  genesisTestTime,
	}, atropos, common.Hash{})
	if err != nil {
		panic(err)
	}
	dbs.Flush(atropos.Bytes())

	input := NewEventStore(nil)

	poset := New(lachesis.FakeNetDagConfig(), store, input)

	extended := &ExtendedPoset{
		Poset:  poset,
		blocks: map[idx.Block]*inter.Block{},
	}

	extended.Bootstrap(func(block *inter.Block, stateHash common.Hash, validators pos.Validators) (common.Hash, pos.Validators) {
		// track block events
		if extended.blocks[block.Index] != nil {
			extended.Log.Crit("Created block twice")
		}
		extended.blocks[block.Index] = block
		return stateHash, validators
	})

	return extended, store, input
}

func flushDb(p *ExtendedPoset, e hash.Event) error {
	return p.store.dbs.Flush(e.Bytes())
}
