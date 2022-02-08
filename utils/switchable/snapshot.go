package switchable

import (
	"sync"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
)

type Snapshot struct {
	kvdb.Snapshot
	mu sync.RWMutex
}

func (s *Snapshot) SwitchTo(snap kvdb.Snapshot) kvdb.Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.Snapshot
	s.Snapshot = snap
	return old
}

func Wrap(snap kvdb.Snapshot) *Snapshot {
	s := &Snapshot{}
	s.SwitchTo(snap)
	return s
}

// Has checks if key is in the exists.
func (s *Snapshot) Has(key []byte) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.Has(key)
}

// Get returns key-value pair by key.
func (s *Snapshot) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.Get(key)
}

func (s *Snapshot) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Snapshot.Release()
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
func (s *Snapshot) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &switchableIterator{
		mu:       &s.mu,
		upd:      &s.Snapshot,
		cur:      s.Snapshot,
		parentIt: s.Snapshot.NewIterator(prefix, start),
		prefix:   prefix,
		start:    start,
	}
}

/*
 * Iterator
 */

type switchableIterator struct {
	mu       *sync.RWMutex
	upd      *kvdb.Snapshot
	cur      kvdb.Snapshot
	parentIt kvdb.Iterator

	prefix, start []byte
	key, value    []byte
}

func (it *switchableIterator) mayReopen() {
	if it.cur != *it.upd {
		// reopen iterator if DB was switched
		it.cur = *it.upd
		if it.key != nil {
			it.start = common.CopyBytes(it.key[len(it.prefix):])
		}
		it.parentIt = it.cur.NewIterator(it.prefix, it.start)
		if it.key != nil {
			_ = it.parentIt.Next() // skip previous key
		}
	}
}

// Next scans key-value pair by key in lexicographic order. Looks in cache first,
// then - in DB.
func (it *switchableIterator) Next() bool {
	it.mu.RLock()
	defer it.mu.RUnlock()

	it.mayReopen()

	ok := it.parentIt.Next()
	if !ok {
		it.key = nil
		it.value = nil
		return false
	}
	it.key = it.parentIt.Key()
	it.value = it.parentIt.Value()
	return true
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error. A memory iterator cannot encounter errors.
func (it *switchableIterator) Error() error {
	it.mu.RLock()
	defer it.mu.RUnlock()

	it.mayReopen()

	return it.parentIt.Error()
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (it *switchableIterator) Key() []byte {
	return it.key
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (it *switchableIterator) Value() []byte {
	return it.value
}

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (it *switchableIterator) Release() {
	it.mu.RLock()
	defer it.mu.RUnlock()

	it.mayReopen()

	it.parentIt.Release()
}
