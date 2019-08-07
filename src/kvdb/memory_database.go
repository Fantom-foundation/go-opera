package kvdb

// MemDatabase is a memory-only database.
// Do not use for any production. It does not get persisted!
type MemDatabase struct {
	*CacheWrapper
}

func (w *MemDatabase) NotFlushedPairs() int {
	return 0
}

func (w *MemDatabase) Flush() error {
	return nil
}

func (w *MemDatabase) ClearNotFlushed() {}

// MemDatabase is a memory-only database.
// Do not use for any production. It does not get persisted!
// Technically, it's the same as CacheWrapper, but it has no parent DB.
func NewMemDatabase() *MemDatabase {
	return &MemDatabase{
		CacheWrapper: NewCacheWrapper(NewEmptyDb()),
	}
}
