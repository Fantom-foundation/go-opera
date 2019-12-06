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
	validators := pos.NewValidators()
	for _, addr := range nodes {
		validators.Set(addr, 1)
	}

	mems := memorydb.NewProducer(namespace, mods...)
	dbs := flushable.NewSyncedPool(mems)
	store := NewStore(dbs, LiteStoreConfig())

	atropos := hash.ZeroEvent
	err := store.ApplyGenesis(&genesis.Genesis{
		Time:       genesisTestTime,
		Validators: *validators,
	}, atropos, common.Hash{})
	if err != nil {
		panic(err)
	}
	_ = dbs.Flush(atropos.Bytes())

	input := NewEventStore(nil)

	config := lachesis.FakeNetDagConfig()
	if config.EpochLen > 100 {
		// EpochLen too big, test timeout is possible.
		config.EpochLen = 100
	}
	poset := New(config, store, input)

	extended := &ExtendedPoset{
		Poset:  poset,
		blocks: map[idx.Block]*inter.Block{},
	}

	extended.Bootstrap(inter.ConsensusCallbacks{
		ApplyBlock: func(block *inter.Block, decidedFrame idx.Frame, cheaters inter.Cheaters) (newAppHash common.Hash, sealEpoch bool) {
			// track block events
			if extended.blocks[block.Index] != nil {
				extended.Log.Crit("Created block twice")
			}
			extended.blocks[block.Index] = block

			return common.Hash{}, false
		},
	})

	return extended, store, input
}

func flushDb(p *ExtendedPoset, e hash.Event) error {
	return p.store.dbs.Flush(e.Bytes())
}
