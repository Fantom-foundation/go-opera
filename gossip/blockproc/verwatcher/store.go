package verwatcher

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/Fantom-foundation/go-opera/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	mainDb kvdb.Store

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(mainDb kvdb.Store) *Store {
	s := &Store{
		mainDb:   mainDb,
		Instance: logger.MakeInstance(),
	}

	return s
}
