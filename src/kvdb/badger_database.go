package kvdb

import (
	"github.com/dgraph-io/badger"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// BadgerDatabase is a kvbd.Database wrapper of *badger.DB
type BadgerDatabase struct {
	db *badger.DB
}

// NewBadgerDatabase wraps *badger.DB
func NewBadgerDatabase(db *badger.DB) *BadgerDatabase {
	return &BadgerDatabase{
		db: db,
	}
}

/*
 * Database interface implementation
 */

// Put puts key-value pair into db.
func (w *BadgerDatabase) Put(key []byte, value []byte) error {
	tx := w.db.NewTransaction(true)
	defer tx.Discard()

	err := tx.Set(key, common.CopyBytes(value))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Has checks if key is in the db.
func (w *BadgerDatabase) Has(key []byte) (bool, error) {
	err := w.db.View(func(txn *badger.Txn) error {
		_, rerr := txn.Get(key)
		return rerr
	})

	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Get returns key-value pair by key.
func (w *BadgerDatabase) Get(key []byte) (res []byte, err error) {
	err = w.db.View(func(txn *badger.Txn) error {
		item, rerr := txn.Get(key)
		if rerr != nil {
			return rerr
		}
		res, rerr = item.ValueCopy(res)
		return rerr
	})

	if err == badger.ErrKeyNotFound {
		err = nil
	}
	return
}

// Delete removes key-value pair by key.
func (w *BadgerDatabase) Delete(key []byte) error {
	tx := w.db.NewTransaction(true)
	defer tx.Discard()

	err := tx.Delete(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	}

	return tx.Commit()
}

// Close leaves underlying database.
func (w *BadgerDatabase) Close() {
	w.db = nil
}

// NewBatch creates new batch.
func (w *BadgerDatabase) NewBatch() Batch {
	return &badgerBatch{db: w}
}

/*
 * Batch
 */

// badgerBatch is a batch structure.
type badgerBatch struct {
	db     *BadgerDatabase
	writes []kv
	size   int
}

// Put puts key-value pair into batch.
func (b *badgerBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

// Delete removes key-value pair from batch by key.
func (b *badgerBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil, true})
	b.size++
	return nil
}

// Write writes batch into db.
func (b *badgerBatch) Write() error {
	tx := b.db.db.NewTransaction(true)
	defer tx.Discard()

	var err error
	for _, kv := range b.writes {
		if kv.del {
			err = tx.Delete(kv.k)
		} else {
			err = tx.Set(kv.k, kv.v)
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ValueSize returns values sizes sum.
func (b *badgerBatch) ValueSize() int {
	return b.size
}

// Reset cleans whole batch.
func (b *badgerBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
