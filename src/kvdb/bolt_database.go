package kvdb

import (
	"go.etcd.io/bbolt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

const defaultBucketName = "lachesis"

// BoltDatabase is a kvbd.Database wrapper of *bbolt.DB
type BoltDatabase struct {
	db   *bbolt.DB
	main []byte
}

// NewBoltDatabase wraps *bbolt.DB
func NewBoltDatabase(db *bbolt.DB) *BoltDatabase {
	bucketName := []byte(defaultBucketName)

	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		panic(err)
	}

	return &BoltDatabase{
		db:   db,
		main: bucketName,
	}
}

/*
 * Database interface implementation
 */

// Put puts key-value pair into batch.
func (w *BoltDatabase) Put(key []byte, value []byte) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.main)

		return bucket.Put(key, value)
	})
}

// Has checks if key is in the db.
func (w *BoltDatabase) Has(key []byte) (exists bool, err error) {
	err = w.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.main)
		exists = bucket.Get(key) != nil

		return nil
	})

	return
}

// Get returns key-value pair by key.
func (w *BoltDatabase) Get(key []byte) (val []byte, err error) {
	err = w.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.main)
		val = common.CopyBytes(bucket.Get(key))

		return nil
	})

	return
}

// Delete removes key-value pair by key.
func (w *BoltDatabase) Delete(key []byte) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(w.main)

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
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

// Delete removes key-value pair from batch by key.
func (b *boltBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil, true})
	b.size++
	return nil
}

// Write writes batch into db.
func (b *boltBatch) Write() error {
	return b.wrapper.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.wrapper.main)

		for _, kv := range b.writes {
			var err error
			if kv.del {
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
