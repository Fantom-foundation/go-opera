package kvdb

import (
	"fmt"

	"github.com/dgraph-io/badger"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

var errEnough = fmt.Errorf("enough")

// BadgerDatabase is a kvbd.Database wrapper of *badger.DB
type BadgerDatabase struct {
	db     *badger.DB
	prefix []byte
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

// NewTable returns a Database object that prefixes all keys with a given prefix.
func (w *BadgerDatabase) NewTable(prefix []byte) Database {
	base := common.CopyBytes(w.prefix)
	return &BadgerDatabase{
		db:     w.db,
		prefix: append(append(base, []byte("-")...), prefix...),
	}
}

func (w *BadgerDatabase) fullKey(key []byte) []byte {
	base := common.CopyBytes(w.prefix)
	return append(append(base, separator...), key...)
}

// Put puts key-value pair into db.
func (w *BadgerDatabase) Put(key []byte, value []byte) error {
	tx := w.db.NewTransaction(true)
	defer tx.Discard()

	key = w.fullKey(key)

	err := tx.Set(key, common.CopyBytes(value))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Has checks if key is in the db.
func (w *BadgerDatabase) Has(key []byte) (bool, error) {
	key = w.fullKey(key)

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
	key = w.fullKey(key)

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

// ForEach scans key-value pair by key prefix.
func (w *BadgerDatabase) ForEach(prefix []byte, do func(key, val []byte) bool) error {
	prefix = w.fullKey(prefix)

	err := w.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			k = k[len(w.prefix)+len(separator):]
			err := item.Value(func(v []byte) error {
				if !do(k, v) {
					return errEnough
				}
				return nil
			})
			if err == errEnough {
				return nil
			}
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// Delete removes key-value pair by key.
func (w *BadgerDatabase) Delete(key []byte) error {
	key = w.fullKey(key)

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

type kv struct {
	k, v []byte
}

// badgerBatch is a batch structure.
type badgerBatch struct {
	db     *BadgerDatabase
	writes []kv
	size   int
}

// Put puts key-value pair into batch.
func (b *badgerBatch) Put(key, value []byte) error {
	key = b.db.fullKey(key)

	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value)})
	b.size += len(value) + len(key)
	return nil
}

// Delete removes key-value pair from batch by key.
func (b *badgerBatch) Delete(key []byte) error {
	key = b.db.fullKey(key)

	b.writes = append(b.writes, kv{common.CopyBytes(key), nil})
	b.size += len(key)
	return nil
}

// Write writes batch into db.
func (b *badgerBatch) Write() error {
	tx := b.db.db.NewTransaction(true)
	defer tx.Discard()

	var err error
	for _, kv := range b.writes {
		if kv.v == nil {
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
