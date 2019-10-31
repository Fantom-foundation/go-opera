package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// SetBlock stores chain block.
func (s *Store) SetBlock(b *inter.Block) {
	s.set(s.table.Blocks, b.Index.Bytes(), b)

	// Add to LRU cache.
	if b != nil && s.cache.Blocks != nil {
		s.cache.Blocks.Add(string(b.Index.Bytes()), b)
	}
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n idx.Block) *inter.Block {
	// Get block from LRU cache first.
	if s.cache.Blocks != nil {
		if c, ok := s.cache.Blocks.Get(string(n.Bytes())); ok {
			if b, ok := c.(*inter.Block); ok {
				return b
			}
		}
	}

	block, _ := s.get(s.table.Blocks, n.Bytes(), &inter.Block{}).(*inter.Block)

	// Add to LRU cache.
	if block != nil && s.cache.Blocks != nil {
		s.cache.Blocks.Add(string(block.Index.Bytes()), block)
	}

	return block
}

// SetBlockIndex stores chain block index.
func (s *Store) SetBlockIndex(id hash.Event, n idx.Block) {
	if err := s.table.BlockHashes.Put(id.Bytes(), n.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetBlockIndex returns stored block index.
func (s *Store) GetBlockIndex(id hash.Event) *idx.Block {
	buf, err := s.table.BlockHashes.Get(id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}
	n := idx.BytesToBlock(buf)
	return &n
}

// SetBlockGasUsed save block used gas
func (s *Store) SetBlockGasUsed(n idx.Block, gas uint64) {
	err := s.table.BlockParticipation.Put(n.Bytes(), bigendian.Int64ToBytes(gas))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}

	s.cache.BlockParticipation.Add(string(n.Bytes()), gas)
}

// GetBlockGasUsed return block used gas
func (s *Store) GetBlockGasUsed(n idx.Block) uint64 {
	gasVal, ok := s.cache.BlockParticipation.Get(string(n.Bytes()))
	if ok {
		gas, ok := gasVal.(uint64)
		if ok {
			return gas
		}
	}

	gasBytes, err := s.table.BlockParticipation.Get(n.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if gasBytes == nil {
		return 0
	}

	gas := bigendian.BytesToInt64(gasBytes)
	s.cache.BlockParticipation.Add(string(n.Bytes()), gas)

	return gas
}
