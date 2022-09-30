package evmstore

import (
	"context"
	"os"
	"path/filepath"

	//"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/Fantom-foundation/go-opera/opera/genesis"

	"github.com/Fantom-foundation/go-opera/erigon"
	"github.com/ledgerwatch/erigon-lib/kv"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(db kv.RwDB, g genesis.Genesis) (err error) {
	//batch := s.EvmDb.NewBatch()

	tx, err := db.BeginRw(context.Background())
	if err != nil {
		return err
	}

	path := filepath.Join(erigon.DefaultDataDir(), "erigon", "batch")
	defer os.RemoveAll(path)
	batch := erigon.NewHashBatch(tx, nil, path)
	defer batch.Rollback()

	g.RawEvmItems.ForEach(func(key, value []byte) bool {
		if err != nil {
			return false
		}
		err = batch.Put(kv.PlainState, key, value)
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	if err := batch.Commit(); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
