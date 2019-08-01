package kvdb

import (
	"bytes"
	"errors"
	"sync"

	rbt "github.com/emirpasic/gods/trees/redblacktree"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// CacheWrapper is a kvdb.Database wrapper around any BatchedDatabase.
// On reading, it looks in memory cache first. If not found, it looks in a parent DB.
// On writing, it writes only in cache. To flush the cache into parent DB, call Flush().
type CacheWrapper struct {
	parent Database
	prefix []byte

	modified *rbt.Tree // modified, comparing to parent, pairs. deleted values are nil

	lock *sync.Mutex // we have no guarantees that rbt.Tree works with concurrent reads, so we can't use MutexRW
}

// NewCacheWrapper wraps parent. All the writes into the cache won't be written in parent until .Flush() is called.
func NewCacheWrapper(parent Database) *CacheWrapper {
	return &CacheWrapper{
		parent:   parent,
		modified: rbt.NewWithStringComparator(),
		lock:     new(sync.Mutex),
	}
}

/*
 * Database interface implementation
 */

// NewTableFlushable returns a Database object that prefixes all keys with a given prefix.
func (w *CacheWrapper) NewTableFlushable(prefix []byte) FlushableDatabase {
	base := common.CopyBytes(w.prefix)
	return &CacheWrapper{
		parent:   w.parent,
		modified: w.modified,
		prefix:   append(append(base, []byte("-")...), prefix...),
		lock:     w.lock,
	}
}

// NewTable returns a Database object that prefixes all keys with a given prefix.
func (w *CacheWrapper) NewTable(prefix []byte) Database {
	return w.NewTableFlushable(prefix)
}

// prefixed key
func (w *CacheWrapper) fullKey(key []byte) []byte {
	base := common.CopyBytes(w.prefix)
	return append(append(base, separator...), key...)
}

// Put puts key-value pair into the cache.
func (w *CacheWrapper) Put(key []byte, value []byte) error {
	if value == nil {
		return errors.New("CacheWrapper: value is nil")
	}

	w.lock.Lock()
	defer w.lock.Unlock()

	key = w.fullKey(key)

	w.modified.Put(string(key), common.CopyBytes(value))
	return nil
}

// Has checks if key is in the exists. Looks in cache first, then - in DB.
func (w *CacheWrapper) Has(key []byte) (bool, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	key = w.fullKey(key)

	val, ok := w.modified.Get(string(key))
	if ok {
		return val != nil, nil
	}
	return w.parent.Has(key)
}

// Get returns key-value pair by key. Looks in cache first, then - in DB.
func (w *CacheWrapper) Get(key []byte) ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	key = w.fullKey(key)

	if entry, ok := w.modified.Get(string(key)); ok {
		if entry == nil {
			return nil, nil
		}
		return common.CopyBytes(entry.([]byte)), nil
	}
	return w.parent.Get(key)
}

// returns the smallest node which is > than specified node
func nextNode(tree *rbt.Tree, node *rbt.Node) (next *rbt.Node, ok bool) {
	origin := node
	if node.Right != nil {
		node = node.Right
		for node.Left != nil {
			node = node.Left
		}
		return node, node != nil
	}
	if node.Parent != nil {
		for node.Parent != nil {
			node = node.Parent
			if tree.Comparator(origin.Key, node.Key) <= 0 {
				return node, node != nil
			}
		}
	}

	return nil, false
}

func castToPair(node *rbt.Node) (key, val []byte) {
	if node == nil {
		return nil, nil
	}
	key = []byte(node.Key.(string))
	if node.Value == nil {
		val = nil // deleted key
	} else {
		val = node.Value.([]byte) // putted value
	}
	return key, val
}

// ForEach scans key-value pair by key prefix in lexicographic order. Looks in cache first, then - in DB.
func (w *CacheWrapper) ForEach(prefix []byte, do func(key, val []byte) bool) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	prefix = w.fullKey(prefix)
	globalCont := true // if false, stop iterating both parent and tree

	// call 'do' if pair qualifies
	doIfSuitable := func(key, val, prevKey []byte) (cont bool, newPrevKey []byte) {
		// if false, stop iterating both parent and tree
		if !globalCont {
			return false, prevKey
		}
		// check that prefixed. stop iterating tree if it is
		if !bytes.HasPrefix(key, prefix) {
			return false, prevKey
		}
		// check that val != nil (it means it's removed in tree). move to next tree's key if it is
		if val == nil {
			return true, key
		}
		// check that parent's key is bigger than prev returned key. move to next key if it is
		if prevKey == nil || bytes.Compare(key, prevKey) > 0 {
			// erase key's prefix, because I'm a table for external world
			globalCont = do(key[len(w.prefix)+len(separator):], val) // if 'do' returned false, then never call it again
			return globalCont, key                                   // next key must be greater
		}
		return true, key
	}

	// read first pair from tree
	treeNode, treeOk := w.modified.Ceiling(string(prefix)) // not strict >=
	treeKey, treeVal := castToPair(treeNode)
	var prevKey []byte

	step := func(parentKey, parentVal []byte) bool {
		// until key from the tree is <= parents's key, use tree's key (because it has priority over parent pairs)
		for treeOk && (parentKey == nil || bytes.Compare(treeKey, parentKey) <= 0) {
			// it's not possible that treeKey isn't bigger than prevKey
			// treeKey may be not prefixed
			// treeVal may be nil (i.e. deleted)
			treeOk, prevKey = doIfSuitable(treeKey, treeVal, prevKey)
			if !treeOk {
				break
			}
			treeNode, treeOk = nextNode(w.modified, treeNode) // strict >
			treeKey, treeVal = castToPair(treeNode)
		}
		if parentKey == nil {
			return false // dummy flag, passed below. means that we shouldn't use parent's pair
		}
		// try to use parents's key

		// it's possible that parentKey is smaller than prevKey
		// parentKey is prefixed
		// parentVal cannot be nil
		// parentKey may be deleted in tree (so we shouldn't use it). but it'll be checked anyway by comparing with prevKey
		var cont bool
		cont, prevKey = doIfSuitable(parentKey, parentVal, prevKey)
		return cont
	}
	// read values from both parent and tree. tree has priority over parent
	err := w.parent.ForEach(prefix, step)
	if err != nil {
		return err
	}
	// read all the left pairs from tree
	if globalCont {
		step(nil, nil)
	}

	return nil
}

// Delete removes key-value pair by key. In parent DB, key won't be deleted until .Flush() is called.
func (w *CacheWrapper) Delete(key []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	key = w.fullKey(key)

	w.modified.Put(string(key), nil)
	return nil
}

// Drop all the not flashed keys. After this call, the state is identical to the state of parent DB.
func (w *CacheWrapper) ClearNotFlushed() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.modified.Clear()
}

// Close leaves underlying database.
func (w *CacheWrapper) Close() {
	w.modified = nil
	w.parent.Close()
}

// NewBatch creates new batch.
func (w *CacheWrapper) NewBatch() Batch {
	return &cacheBatch{db: w}
}

// Num of not flushed keys, including deleted keys.
func (w *CacheWrapper) NotFlushedPairs() int {
	return w.modified.Size()
}

// Flushes current cache into parent DB.
func (w *CacheWrapper) Flush() error {
	batch := w.parent.NewBatch()
	for it := w.modified.Iterator(); it.Next(); {
		var err error

		if it.Value() == nil {
			err = batch.Delete([]byte(it.Key().(string)))
		} else {
			err = batch.Put([]byte(it.Key().(string)), it.Value().([]byte))
		}

		if err != nil {
			return err
		}

		if batch.ValueSize() > IdealBatchSize {
			err = batch.Write()
			if err != nil {
				return err
			}
			batch.Reset()
		}
	}
	w.modified.Clear()

	return batch.Write()
}

/*
 * Batch
 */

// cacheBatch is a batch structure.
type cacheBatch struct {
	db     Database
	writes []kv
	size   int
}

// Put puts key-value pair into batch.
func (b *cacheBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value)})
	b.size += len(value) + len(key)
	return nil
}

// Delete removes key-value pair from batch by key.
func (b *cacheBatch) Delete(key []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), nil})
	b.size += len(key)
	return nil
}

// Write writes batch into db. Not atomic.
func (b *cacheBatch) Write() error {
	for _, kv := range b.writes {
		var err error

		if kv.v == nil {
			err = b.db.Delete(kv.k)
		} else {
			err = b.db.Put(kv.k, kv.v)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// ValueSize returns values sizes sum.
func (b *cacheBatch) ValueSize() int {
	return b.size
}

// Reset cleans whole batch.
func (b *cacheBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
