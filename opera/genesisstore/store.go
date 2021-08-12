package genesisstore

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	db kvdb.Store

	table struct {
		Rules kvdb.Store `table:"c"`

		Blocks kvdb.Store `table:"b"`

		EvmAccounts kvdb.Store `table:"a"`
		EvmStorage  kvdb.Store `table:"s"`
		RawEvmItems kvdb.Store `table:"M"`

		Delegations kvdb.Store `table:"d"`
		Metadata    kvdb.Store `table:"m"`
	}

	rlp rlpstore.Helper
	logger.Instance
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	return NewStore(memorydb.New())
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Store) *Store {
	s := &Store{
		db:       db,
		Instance: logger.New("genesis-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
	}

	table.MigrateTables(&s.table, s.db)

	return s
}

// Close leaves underlying database.
func (s *Store) Close() {
	table.MigrateTables(&s.table, nil)

	_ = s.db.Close()
}
