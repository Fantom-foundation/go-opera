package table

import (
	"bytes"

	"github.com/ethereum/go-ethereum/ethdb"
)

// Table wraps the underling DB, so all the table's data is stored with a prefix in underling DB
type Table struct {
	db     ethdb.KeyValueStore
	prefix []byte
}

var (
	// NOTE: key collisions are possible
	separator = []byte{}
)

// prefixed key (prefix + separator + key)
func prefixed(key, prefix []byte) []byte {
	prefixedKey := make([]byte, 0, len(prefix)+len(separator)+len(key))
	prefixedKey = append(prefixedKey, prefix...)
	prefixedKey = append(prefixedKey, separator...)
	prefixedKey = append(prefixedKey, key...)
	return prefixedKey
}

func noPrefix(key, prefix []byte) []byte {
	if len(key) < len(prefix)+len(separator) {
		return key
	}
	return key[len(prefix)+len(separator):]
}

/*
 * Database
 */

func New(db ethdb.KeyValueStore, prefix []byte) *Table {
	return &Table{db, prefix}
}

func (t Table) NewTable(prefix []byte) *Table {
	return &Table{t.db, prefix}
}

func (t *Table) Close() error {
	return nil
}

// Drop the whole database.
func (t *Table) Drop() {}

func (t *Table) Has(key []byte) (bool, error) {
	return t.db.Has(prefixed(key, t.prefix))
}

func (t *Table) Get(key []byte) ([]byte, error) {
	return t.db.Get(prefixed(key, t.prefix))
}

func (t *Table) Put(key []byte, value []byte) error {
	return t.db.Put(prefixed(key, t.prefix), value)
}

func (t *Table) Delete(key []byte) error {
	return t.db.Delete(prefixed(key, t.prefix))
}

func (t *Table) NewBatch() ethdb.Batch {
	return &batch{t.db.NewBatch(), t.prefix}
}

func (t *Table) Stat(property string) (string, error) {
	return t.db.Stat(property)
}

func (t *Table) Compact(start []byte, limit []byte) error {
	return t.db.Compact(start, limit)
}

/*
 * Iterator
 */

type iterator struct {
	it     ethdb.Iterator
	prefix []byte
}

func (it *iterator) Next() bool {
	next := it.it.Next()
	for next && !bytes.HasPrefix(it.it.Key(), it.prefix) {
		next = it.it.Next()
	}
	return next
}

func (it *iterator) Error() error {
	return it.it.Error()
}

func (it *iterator) Key() []byte {
	return noPrefix(it.it.Key(), it.prefix)
}

func (it *iterator) Value() []byte {
	return it.it.Value()
}

func (it *iterator) Release() {
	it.it.Release()
	*it = iterator{}
}

func (t *Table) NewIterator() ethdb.Iterator {
	return &iterator{t.db.NewIteratorWithPrefix(t.prefix), t.prefix}
}

func (t *Table) NewIteratorWithStart(start []byte) ethdb.Iterator {
	return &iterator{t.db.NewIteratorWithStart(prefixed(start, t.prefix)), t.prefix}
}

func (t *Table) NewIteratorWithPrefix(itPrefix []byte) ethdb.Iterator {
	return &iterator{t.db.NewIteratorWithPrefix(prefixed(itPrefix, t.prefix)), t.prefix}
}

/*
 * Batch
 */

type batch struct {
	batch  ethdb.Batch
	prefix []byte
}

func (b *batch) Put(key, value []byte) error {
	return b.batch.Put(prefixed(key, b.prefix), value)
}

func (b *batch) Delete(key []byte) error {
	return b.batch.Delete(prefixed(key, b.prefix))
}

func (b *batch) ValueSize() int {
	return b.batch.ValueSize()
}

func (b *batch) Write() error {
	return b.batch.Write()
}

func (b *batch) Reset() {
	b.batch.Reset()
}

func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	return b.batch.Replay(&replayer{w, b.prefix})
}

/*
 * Replayer
 */

type replayer struct {
	writer ethdb.KeyValueWriter
	prefix []byte
}

func (r *replayer) Put(key, value []byte) error {
	return r.writer.Put(noPrefix(key, r.prefix), value)
}

func (r *replayer) Delete(key []byte) error {
	return r.writer.Delete(noPrefix(key, r.prefix))
}
