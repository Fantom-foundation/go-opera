package kvdb

import (
	"bytes"

	"go.etcd.io/bbolt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

var (
	// NOTE: key collisions are possible
	separator = []byte("::")
)

// BoltDatabase is a kvbd.Database wrapper of *bbolt.DB.
type BoltDatabase struct {
	db     *bbolt.DB
	bucket []byte
}

func newBoltDatabase(db *bbolt.DB, bucket []byte) *BoltDatabase {
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		panic(err)
	}

	return &BoltDatabase{
		db:     db,
		bucket: bucket,
	}
}

// NewBoltDatabase wraps *bbolt.DB.
func NewBoltDatabase(db *bbolt.DB) *BoltDatabase {
	return newBoltDatabase(db, []byte("default"))
}

/*
 * Database interface implementation
 */

// NewTable returns a Database object that prefixes all keys with a given prefix.
func (w *BoltDatabase) NewTable(prefix []byte) Database {
	return newBoltDatabase(w.db, prefix)
}

// Put puts key-value pair into batch.
func (w *BoltDatabase) Put(key []byte, value []byte) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.bucket)

		return bucket.Put(key, value)
	})
}

// Has checks if key is in the db.
func (w *BoltDatabase) Has(key []byte) (exists bool, err error) {
	err = w.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.bucket)
		exists = bucket.Get(key) != nil

		return nil
	})

	return
}

// Get returns key-value pair by key.
func (w *BoltDatabase) Get(key []byte) (val []byte, err error) {
	err = w.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.bucket)
		val = common.CopyBytes(bucket.Get(key))

		return nil
	})

	return
}

// ForEach scans key-value pair by key prefix.
func (w *BoltDatabase) ForEach(prefix []byte, do func(key, val []byte) bool) error {
	err := w.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(w.bucket).Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if !do(k, v) {
				return nil
			}
		}
		return nil
	})

	return err
}

// Delete removes key-value pair by key.
func (w *BoltDatabase) Delete(key []byte) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.bucket)

		return bucket.Delete(key)
	})
}

// Close leaves underlying database.
func (w *BoltDatabase) Close() {
	w.db = nil
}

// NewBatch creates new batch.
func (w *BoltDatabase) NewBatch() Batch {
	return &boltBatch{wrapper: w}
}

/*
 * Batch
 */

// boltBatch is a batch structure
type boltBatch struct {
	wrapper *BoltDatabase
	writes  []kv
	size    int
}

// Put puts key-value pair into batch.
func (b *boltBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value)})
	b.size += len(value) + len(key)
	return nil
}

// Delete removes key-value pair from batch by key.
func (b *boltBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil})
	b.size += len(key)
	return nil
}

// Write writes batch into db.
func (b *boltBatch) Write() error {
	return b.wrapper.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.wrapper.bucket)

		for _, kv := range b.writes {
			var err error
			if kv.v == nil {
				err = bucket.Delete(kv.k)
			} else {
				err = bucket.Put(kv.k, kv.v)
			}
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// ValueSize returns values sizes sum.
func (b *boltBatch) ValueSize() int {
	return b.size
}

// Reset cleans whole batch.
func (b *boltBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
