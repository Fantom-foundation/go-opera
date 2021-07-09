package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
)

func (s *Store) GetGenesisHash() *hash.Hash {
	valBytes, err := s.table.Genesis.Get([]byte("g"))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if len(valBytes) == 0 {
		return nil
	}
	val := hash.BytesToHash(valBytes)
	return &val
}

func (s *Store) SetGenesisHash(val hash.Hash) {
	err := s.table.Genesis.Put([]byte("g"), val.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// SetBlock stores chain block.
func (s *Store) SetBlock(n idx.Block, b *inter.Block) {
	s.rlp.Set(s.table.Blocks, n.Bytes(), b)

	// Add to LRU cache.
	s.cache.Blocks.Add(n, b, uint(b.EstimateSize()))
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n idx.Block) *inter.Block {
	// Get block from LRU cache first.
	if c, ok := s.cache.Blocks.Get(n); ok {
		return c.(*inter.Block)
	}

	block, _ := s.rlp.Get(s.table.Blocks, n.Bytes(), &inter.Block{}).(*inter.Block)

	// Add to LRU cache.
	if block != nil {
		s.cache.Blocks.Add(n, block, uint(block.EstimateSize()))
	}

	return block
}

func (s *Store) ForEachBlock(fn func(index idx.Block, block *inter.Block)) {
	it := s.table.Blocks.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		var block inter.Block
		err := rlp.DecodeBytes(it.Value(), &block)
		if err != nil {
			s.Log.Crit("Failed to decode block", "err", err)
		}
		fn(idx.BytesToBlock(it.Key()), &block)
	}
}

// SetBlockIndex stores chain block index.
func (s *Store) SetBlockIndex(id hash.Event, n idx.Block) {
	if err := s.table.BlockHashes.Put(id.Bytes(), n.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}

	s.cache.BlockHashes.Add(id, n, nominalSize)
}

// GetBlockIndex returns stored block index.
func (s *Store) GetBlockIndex(id hash.Event) *idx.Block {
	nVal, ok := s.cache.BlockHashes.Get(id)
	if ok {
		n, ok := nVal.(idx.Block)
		if ok {
			return &n
		}
	}

	buf, err := s.table.BlockHashes.Get(id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}
	n := idx.BytesToBlock(buf)

	s.cache.BlockHashes.Add(id, n, nominalSize)

	return &n
}

// SetGenesisBlockIndex stores genesis block index.
func (s *Store) SetGenesisBlockIndex(n idx.Block) {
	if err := s.table.Genesis.Put([]byte("i"), n.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetGenesisBlockIndex returns stored genesis block index.
func (s *Store) GetGenesisBlockIndex() *idx.Block {
	buf, err := s.table.Genesis.Get([]byte("i"))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}
	n := idx.BytesToBlock(buf)

	return &n
}

func (s *Store) GetGenesisTime() inter.Timestamp {
	n := s.GetGenesisBlockIndex()
	if n == nil {
		return 0
	}
	block := s.GetBlock(*n)
	if block == nil {
		return 0
	}
	return block.Time
}
