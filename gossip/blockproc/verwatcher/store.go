package verwatcher

import (
	"sync/atomic"

	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/Fantom-foundation/go-opera/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	mainDB kvdb.Store

	cache struct {
		networkVersion atomic.Value
		missedVersion  atomic.Value
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store) *Store {
	s := &Store{
		mainDB:   mainDB,
		Instance: logger.New("verwatcher-store"),
	}

	return s
}
