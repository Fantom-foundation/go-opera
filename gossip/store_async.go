package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
)

type asyncStore struct {
	dbs    *flushable.SyncedPool
	mainDb kvdb.Store
	table  struct {
		// Network tables
		Peers kvdb.Store `table:"Z"`
	}
}

func newAsyncStore(dbs *flushable.SyncedPool) *asyncStore {
	s := &asyncStore{
		dbs:    dbs,
		mainDb: dbs.GetDb("gossip-async"),
	}

	table.MigrateTables(&s.table, s.mainDb)

	return s
}

// Close leaves underlying database.
func (s *asyncStore) Close() {
	table.MigrateTables(&s.table, nil)

	s.mainDb.Close()
}
