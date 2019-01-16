package kvdb

import (
	"errors"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// MemDatabase is a kvbd.Database wrapper of map[string][]byte
// Do not use for any production it does not get persisted
type MemDatabase struct {
	db   map[string][]byte
	lock sync.RWMutex
}

// NewMemDatabase wraps map[string][]byte
func NewMemDatabase() *MemDatabase {
	return &MemDatabase{
		db: make(map[string][]byte),
	}
}

// Keys returns slice of keys in the db.
func (w *MemDatabase) Keys() [][]byte {
	w.lock.RLock()
	defer w.lock.RUnlock()

	keys := [][]byte{}
	for key := range w.db {
		keys = append(keys, []byte(key))
	}
	return keys
}

// Len returns count of key-values pairs in the db.
func (w *MemDatabase) Len() int {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return len(w.db)
}

/*
 * Database interface implementation
 */

// Put puts key-value pair into db.
func (w *MemDatabase) Put(key []byte, value []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.db[string(key)] = common.CopyBytes(value)
	return nil
}

// Has checks if key is in the db.
func (w *MemDatabase) Has(key []byte) (bool, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()

	_, ok := w.db[string(key)]
	return ok, nil
}

// Get returns key-value pair by key.
func (w *MemDatabase) Get(key []byte) ([]byte, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()

	if entry, ok := w.db[string(key)]; ok {
		return common.CopyBytes(entry), nil
	}
	return nil, errors.New("not found")
}

// Delete removes key-value pair by key.
func (w *MemDatabase) Delete(key []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	delete(w.db, string(key))
	return nil
}

// Close does nothing.
func (w *MemDatabase) Close() {}

// NewBatch creates new batch.
func (w *MemDatabase) NewBatch() Batch {
	return &memBatch{db: w}
}

/*
 * Batch
 */

type kv struct {
	k, v []byte
	del  bool
}

// memBatch is a batch structure.
type memBatch struct {
	db     *MemDatabase
	writes []kv
	size   int
}

// Put puts key-value pair into batch.
func (b *memBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

// Delete removes key-value pair from batch by key.
func (b *memBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil, true})
	b.size++
	return nil
}

// Write writes batch into db.
func (b *memBatch) Write() error {
	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	for _, kv := range b.writes {
		if kv.del {
			delete(b.db.db, string(kv.k))
			continue
		}
		b.db.db[string(kv.k)] = kv.v
	}
	return nil
}

// ValueSize returns values sizes sum.
func (b *memBatch) ValueSize() int {
	return b.size
}

// Reset cleans whole batch.
func (b *memBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
