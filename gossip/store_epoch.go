package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"errors"
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/skiperrors"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
)

type (
	epochStore struct {
		epoch idx.Epoch
		db    kvdb.DropableStore
		table struct {
			Tips     kvdb.Store `table:"t"`
			Heads    kvdb.Store `table:"H"`
			DagIndex kvdb.Store `table:"v"`
		}
	}
)

func newEpochStore(epoch idx.Epoch, db kvdb.DropableStore) *epochStore {
	es := &epochStore{
		epoch: epoch,
		db:    db,
	}
	table.MigrateTables(&es.table, db)

	err := errors.New("database closed")

	// wrap with skiperrors to skip errors on reading from a dropped DB
	es.table.Tips = skiperrors.Wrap(es.table.Tips, err)
	es.table.Heads = skiperrors.Wrap(es.table.Heads, err)

	return es
}

func (s *Store) getAnyEpochStore() *epochStore {
	_es := s.epochStore.Load()
	if _es == nil {
		return nil
	}
	es := _es.(*epochStore)
	return es
}

// getEpochStore is safe for concurrent use.
func (s *Store) getEpochStore(epoch idx.Epoch) *epochStore {
	es := s.getAnyEpochStore()
	if es.epoch != epoch {
		return nil
	}
	return es
}

func (s *Store) resetEpochStore(newEpoch idx.Epoch) {
	oldEs := s.epochStore.Load()
	// create new DB
	s.createEpochStore(newEpoch)
	// drop previous DB
	// there may be race condition with threads which hold this DB, so wrap tables with skiperrors
	if oldEs != nil {
		err := oldEs.(*epochStore).db.Close()
		if err != nil {
			s.Log.Error("Failed to close epoch DB", "err", err)
			return
		}
		oldEs.(*epochStore).db.Drop()
	}
}

func (s *Store) loadEpochStore(epoch idx.Epoch) {
	if s.epochStore.Load() != nil {
		return
	}
	s.createEpochStore(epoch)
}

func (s *Store) closeEpochStore() error {
	es := s.getAnyEpochStore()
	if es == nil {
		return nil
	}
	return es.db.Close()
}

func (s *Store) createEpochStore(epoch idx.Epoch) {
	// create new DB
	name := fmt.Sprintf("gossip-%d", epoch)
	db, err := s.dbs.OpenDB(name)
	if err != nil {
		s.Log.Crit("Filed to open DB", "name", name, "err", err)
	}
	s.epochStore.Store(newEpochStore(epoch, db))
}

// SetLastEvent stores last unconfirmed event from a validator (off-chain)
func (s *Store) SetLastEvent(epoch idx.Epoch, from idx.ValidatorID, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := from.Bytes()
	if err := es.table.Tips.Put(key, id.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetLastEvent returns stored last unconfirmed event from a validator (off-chain)
func (s *Store) GetLastEvent(epoch idx.Epoch, from idx.ValidatorID) *hash.Event {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	key := from.Bytes()
	idBytes, err := es.table.Tips.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if idBytes == nil {
		return nil
	}
	id := hash.BytesToEvent(idBytes)
	return &id
}
