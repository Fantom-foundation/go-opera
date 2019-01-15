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

func (w *BadgerDatabase) Put(key []byte, value []byte) error {
	tx := w.db.NewTransaction(true)
	defer tx.Discard()

	err := tx.Set(key, common.CopyBytes(value))
	if err != nil {
		return err
	}

	return tx.Commit(nil)
}

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

func (w *BadgerDatabase) Get(key []byte) (res []byte, err error) {
	err = w.db.View(func(txn *badger.Txn) error {
		item, rerr := txn.Get(key)
		if rerr != nil {
			return rerr
		}
		res, rerr = item.ValueCopy(res)
		return rerr
	})

	return
}

func (w *BadgerDatabase) Delete(key []byte) error {
	tx := w.db.NewTransaction(true)
	defer tx.Discard()

	return tx.Commit(nil)
}

func (w *BadgerDatabase) Close() {}

func (w *BadgerDatabase) NewBatch() Batch {
	return &badgerBatch{db: w}
}

/*
 * Batch
 */

type badgerBatch struct {
	db     *BadgerDatabase
	writes []kv
	size   int
}

func (b *badgerBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

func (b *badgerBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil, true})
	b.size += 1
	return nil
}

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

	return tx.Commit(nil)
}

func (b *badgerBatch) ValueSize() int {
	return b.size
}

func (b *badgerBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
