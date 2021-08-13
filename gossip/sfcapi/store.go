package sfcapi

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	mainDB kvdb.Store
	table  struct {
		GasPowerRefund kvdb.Store `table:"R"`

		Validators  kvdb.Store `table:"1"`
		Stakers     kvdb.Store `table:"2"`
		Delegations kvdb.Store `table:"3"`

		DelegationOldRewards        kvdb.Store `table:"6"`
		StakerOldRewards            kvdb.Store `table:"7"`
		StakerDelegationsOldRewards kvdb.Store `table:"8"`
	}

	rlp rlpstore.Helper

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store) *Store {
	s := &Store{
		mainDB:   mainDB,
		Instance: logger.New("sfcapi-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
	}

	table.MigrateTables(&s.table, s.mainDB)

	return s
}

// Close closes underlying database.
func (s *Store) Close() {
	table.MigrateTables(&s.table, nil)

	_ = s.mainDB.Close()
}
