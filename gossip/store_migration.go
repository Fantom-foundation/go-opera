package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/Fantom-foundation/go-opera/utils/migration"
)

func isEmptyDB(db kvdb.Iteratee) bool {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	return !it.Next()
}

func (s *Store) Migrate() error {
	versions := migration.NewKvdbIDStore(s.table.Version)
	if isEmptyDB(s.mainDB) && isEmptyDB(s.async.mainDB) {
		// short circuit if empty DB
		versions.SetID(s.migrations().ID())
		return nil
	}
	return s.migrations().Exec(versions)
}

func (s *Store) migrations() *migration.Migration {
	return migration.
		Begin("opera-gossip-store")
}
