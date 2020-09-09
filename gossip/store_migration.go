package gossip

import (
	"github.com/Fantom-foundation/go-opera/utils/migration"
)

func (s *Store) Migrate() error {
	versions := migration.NewKvdbIDStore(s.table.Version)
	return s.migrations().Exec(versions)
}

func (s *Store) migrations() *migration.Migration {
	return migration.
		Begin("opera-gossip-store")
}
