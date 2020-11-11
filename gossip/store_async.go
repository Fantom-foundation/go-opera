package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
)

type asyncStore struct {
	dbs    *flushable.SyncedPool
	mainDB kvdb.Store
	table  struct {
		// Network tables
		Peers kvdb.Store `table:"Z"`
	}
}

func newAsyncStore(dbs *flushable.SyncedPool) *asyncStore {
	s := &asyncStore{
		dbs:    dbs,
		mainDB: dbs.GetDb("gossip-async"),
	}

	table.MigrateTables(&s.table, s.mainDB)

	return s
}

// Close leaves underlying database.
func (s *asyncStore) Close() {
	table.MigrateTables(&s.table, nil)

	s.mainDB.Close()
}
