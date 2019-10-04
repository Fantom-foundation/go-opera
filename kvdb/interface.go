package kvdb

import (
	"github.com/ethereum/go-ethereum/ethdb"
)

// Droper wraps the Drop method of a backing data store.
type Droper interface {
	Drop()
}

// KeyValueStore contains all the methods required to allow handling different
// key-value data stores backing the high level database.
type KeyValueStore interface {
	ethdb.KeyValueStore
	Droper
}

// FlushableKeyValueStore
type FlushableKeyValueStore interface {
	KeyValueStore

	NotFlushedPairs() int
	NotFlushedSizeEst() int
	Flush() error
	DropNotFlushed()
}

// DbProducer represents real db producer.
type DbProducer interface {
	Names() []string

	OpenDb(name string) KeyValueStore
}
