package skiperrors

import (
	"github.com/ethereum/go-ethereum/ethdb"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

// wrapper is a kvdb.KeyValueStore wrapper around any kvdb.KeyValueStore.
// It ignores some errors of underlying store.
// NOTE: ignoring is not implemented at Iterator, Batch, .
type wrapper struct {
	underlying kvdb.KeyValueStore

	errs []error
}

// Wrap returns a wrapped kvdb.KeyValueStore.
func Wrap(db kvdb.KeyValueStore, errs ...error) kvdb.KeyValueStore {
	return &wrapper{
		underlying: db,
		errs:       errs,
	}
}

func (f *wrapper) skip(got error) bool {
	if got == nil {
		return false
	}

	for _, exp := range f.errs {
		if got == exp || got.Error() == exp.Error() {
			return true
		}
	}

	return false
}

/*
 * implementation:
 */

// Has retrieves if a key is present in the key-value data store.
func (f *wrapper) Has(key []byte) (bool, error) {
	has, err := f.underlying.Has(key)
	if f.skip(err) {
		return false, nil
	}
	return has, err
}

// Get retrieves the given key if it's present in the key-value data store.
func (f *wrapper) Get(key []byte) ([]byte, error) {
	b, err := f.underlying.Get(key)
	if f.skip(err) {
		return nil, nil
	}
	return b, err
}

// Put inserts the given value into the key-value data store.
func (f *wrapper) Put(key []byte, value []byte) error {
	err := f.underlying.Put(key, value)
	if f.skip(err) {
		return nil
	}
	return err
}

// Delete removes the key from the key-value data store.
func (f *wrapper) Delete(key []byte) error {
	err := f.underlying.Delete(key)
	if f.skip(err) {
		return nil
	}
	return err
}

// NewBatch creates a write-only database that buffers changes to its host db
// until a final write is called.
func (f *wrapper) NewBatch() ethdb.Batch {
	return f.underlying.NewBatch()
}

// NewIterator creates a binary-alphabetical iterator over the entire keyspace
// contained within the key-value database.
func (f *wrapper) NewIterator() ethdb.Iterator {
	return f.underlying.NewIterator()
}

// NewIteratorWithStart creates a binary-alphabetical iterator over a subset of
// database content starting at a particular initial key (or after, if it does
// not exist).
func (f *wrapper) NewIteratorWithStart(start []byte) ethdb.Iterator {
	return f.underlying.NewIteratorWithStart(start)
}

// NewIteratorWithPrefix creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix.
func (f *wrapper) NewIteratorWithPrefix(prefix []byte) ethdb.Iterator {
	return f.underlying.NewIteratorWithPrefix(prefix)
}

// Stat returns a particular internal stat of the database.
func (f *wrapper) Stat(property string) (string, error) {
	stat, err := f.underlying.Stat(property)
	if f.skip(err) {
		return "", nil
	}
	return stat, err
}

// Compact flattens the underlying data store for the given key range. In essence,
// deleted and overwritten versions are discarded, and the data is rearranged to
// reduce the cost of operations needed to access them.
//
// A nil start is treated as a key before all keys in the data store; a nil limit
// is treated as a key after all keys in the data store. If both is nil then it
// will compact entire data store.
func (f *wrapper) Compact(start []byte, limit []byte) error {
	err := f.underlying.Compact(start, limit)
	if f.skip(err) {
		return nil
	}
	return err
}

// Close closes database.
func (f *wrapper) Close() error {
	err := f.underlying.Close()
	if f.skip(err) {
		return nil
	}
	return err
}

// Drop drops database.
func (f *wrapper) Drop() {
	f.underlying.Drop()
}
