package flushable

import (
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/devnulldb"
)

// LazyFlushable is a Flushable with delayed DB producer
type LazyFlushable struct {
	*Flushable
	producer func() kvdb.KeyValueStore
}

var (
	devnull = devnulldb.New()
)

// NewLazy makes flushable with real db producer.
// Real db won't be produced until first .Flush() is called.
// All the writes into the cache won't be written in parent until .Flush() is called.
func NewLazy(producer func() kvdb.KeyValueStore, drop func()) *LazyFlushable {
	if producer == nil {
		panic("nil producer")
	}

	w := &LazyFlushable{
		Flushable: WrapWithDrop(devnull, drop),
		producer:  producer,
	}
	return w
}

// InitUnderlyingDb is UnderlyingDb getter. Makes underlying in lazy case.
func (w *LazyFlushable) InitUnderlyingDb() kvdb.KeyValueStore {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.initUnderlyingDb()
}

func (w *LazyFlushable) initUnderlyingDb() kvdb.KeyValueStore {
	if w.underlying == devnull && w.producer != nil {
		w.underlying = w.producer()
		w.producer = nil // need once
	}

	return w.underlying
}

// Flush current cache into parent DB.
// Real db won't be produced until first .Flush() is called.
func (w *LazyFlushable) Flush() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.underlying = w.initUnderlyingDb()

	return w.flush()
}
