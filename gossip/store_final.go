package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/log"
)

// OnlyFinal data store.
func (s *Store) OnlyFinal() *Store {
	const name1 = "gossip"
	mainDB, err := s.dbs.GetUnderlying(name1)
	if err != nil {
		log.Crit("fail to open db", "name", name1, "err", err)
	}

	const name2 = "gossip-async"
	asyncDB, err := s.dbs.GetUnderlying("gossip-async")
	if err != nil {
		log.Crit("failed to open db", "name", "err", err)
	}

	return newStore(nil, s.cfg, &readonly{mainDB}, &readonly{asyncDB})
}

type readonly struct {
	kvdb.ReadonlyStore
}

// Put puts key-value pair into the cache.
func (ro *readonly) Put(key []byte, value []byte) error {
	panic("readonly!")
	return nil
}

// Delete removes key-value pair by key.
func (ro *readonly) Delete(key []byte) error {
	panic("readonly!")
	return nil
}

// Compact flattens the underlying data store for the given key range.
func (ro *readonly) Compact(start []byte, limit []byte) error {
	panic("readonly!")
	return nil
}

// Close leaves underlying database.
func (ro *readonly) Close() error {
	panic("readonly!")
	return nil
}

// NewBatch creates new batch.
func (ro *readonly) NewBatch() kvdb.Batch {
	panic("readonly!")
	return nil
}
