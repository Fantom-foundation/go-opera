package kvdb

import (
	"github.com/ethereum/go-ethereum/ethdb"
)

type Droper interface {
	Drop()
}

// KeyValueStore contains all the methods required to allow handling different
// key-value data stores backing the high level database.
type KeyValueStore interface {
	ethdb.KeyValueStore
	Droper
}

type FlushableKeyValueStore interface {
	KeyValueStore

	NotFlushedPairs() int
	NotFlushedSizeEst() int
	Flush() error
	DropNotFlushed()
}
