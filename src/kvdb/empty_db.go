package kvdb

type EmptyDb struct{}

// NewEmptyDb wraps map[string][]byte
func NewEmptyDb() *EmptyDb {
	return &EmptyDb{}
}

/*
 * Database interface implementation
 */

func (w *EmptyDb) NewTable(prefix []byte) Database {
	return w
}

func (w *EmptyDb) Put(key []byte, value []byte) error {
	return nil
}

func (w *EmptyDb) Has(key []byte) (bool, error) {
	return false, nil
}

func (w *EmptyDb) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (w *EmptyDb) ForEach(prefix []byte, do func(key, val []byte) bool) error {
	return nil
}

func (w *EmptyDb) Delete(key []byte) error {
	return nil
}

func (w *EmptyDb) Close() {}

func (w *EmptyDb) NewBatch() Batch {
	return &emptyBatch{}
}

/*
 * Batch
 */

type emptyBatch struct{}

func (b *emptyBatch) Put(key, value []byte) error {
	return nil
}

func (b *emptyBatch) Delete(key []byte) error {
	return nil
}

func (b *emptyBatch) Write() error {
	return nil
}

func (b *emptyBatch) ValueSize() int {
	return 0
}

func (b *emptyBatch) Reset() {}
