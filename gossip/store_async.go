package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/log"
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
	const name = "gossip-async"
	mainDB, err := dbs.OpenDB(name)
	if err != nil {
		log.Crit("failed to open db", "name", name, "err", err)
	}

	s := &asyncStore{
		dbs:    dbs,
		mainDB: mainDB,
	}

	table.MigrateTables(&s.table, s.mainDB)

	return s
}

// Close leaves underlying database.
func (s *asyncStore) Close() {
	table.MigrateTables(&s.table, nil)

	s.mainDB.Close()
}
