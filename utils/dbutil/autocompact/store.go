package autocompact

import (
	"bytes"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
)

type Store struct {
	kvdb.Store
	minKey  []byte
	maxKey  []byte
	written uint64
	limit   uint64
	compMu  sync.Mutex
}

type Batch struct {
	kvdb.Batch
	written uint64
	minKey  []byte
	maxKey  []byte
	onWrite func(key []byte, size uint64, force bool)
}

func Wrap(s kvdb.Store, limit uint64) *Store {
	return &Store{
		Store: s,
		limit: limit,
	}
}

func (s *Store) onWrite(key []byte, size uint64, force bool) {
	s.compMu.Lock()
	defer s.compMu.Unlock()
	s.written += size
	if s.minKey == nil {
		s.minKey = common.CopyBytes(key)
		s.maxKey = common.CopyBytes(key)
	}
	if key != nil && bytes.Compare(key, s.minKey) < 0 {
		s.minKey = common.CopyBytes(key)
	}
	if key != nil && bytes.Compare(key, s.maxKey) > 0 {
		s.maxKey = common.CopyBytes(key)
	}
	if force || s.written > s.limit {
		_ = s.Store.Compact(s.minKey, s.maxKey)
		s.written = 0
		s.minKey = nil
		s.maxKey = nil
	}
}

func (s *Store) Put(key []byte, value []byte) error {
	defer s.onWrite(key, uint64(len(key)+len(value)+64), false)
	return s.Store.Put(key, value)
}

func (s *Store) Delete(key []byte) error {
	defer s.onWrite(key, uint64(len(key)+64), false)
	return s.Store.Delete(key)
}

func (s *Store) Close() error {
	s.onWrite(nil, 0, true)
	return s.Store.Close()
}

func (s *Store) NewBatch() kvdb.Batch {
	batch := s.Store.NewBatch()
	if batch == nil {
		return nil
	}
	return &Batch{
		Batch:   batch,
		onWrite: s.onWrite,
	}
}

func (s *Batch) Put(key []byte, value []byte) error {
	s.written += uint64(len(key) + len(value) + 64)
	if s.minKey == nil {
		s.minKey = common.CopyBytes(key)
		s.maxKey = common.CopyBytes(key)
	}
	if bytes.Compare(key, s.minKey) < 0 {
		s.minKey = common.CopyBytes(key)
	}
	if bytes.Compare(key, s.maxKey) > 0 {
		s.maxKey = common.CopyBytes(key)
	}
	return s.Batch.Put(key, value)
}

func (s *Batch) Delete(key []byte) error {
	s.written += uint64(len(key) + 64)
	if s.minKey == nil {
		s.minKey = common.CopyBytes(key)
		s.maxKey = common.CopyBytes(key)
	}
	if bytes.Compare(key, s.minKey) < 0 {
		s.minKey = common.CopyBytes(key)
	}
	if bytes.Compare(key, s.maxKey) > 0 {
		s.maxKey = common.CopyBytes(key)
	}
	return s.Batch.Delete(key)
}

func (s *Batch) Reset() {
	s.written = 0
	s.minKey = nil
	s.maxKey = nil
	s.Batch.Reset()
}

func (s *Batch) Write() error {
	s.onWrite(s.minKey, 0, false)
	defer s.onWrite(s.maxKey, s.written, false)
	return s.Batch.Write()
}
