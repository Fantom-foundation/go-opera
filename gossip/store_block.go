package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetBlock stores chain block.
func (s *Store) SetBlock(b *inter.Block) {
	s.set(s.table.Blocks, b.Index.Bytes(), b)
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n idx.Block) *inter.Block {
	block, _ := s.get(s.table.Blocks, n.Bytes(), &inter.Block{}).(*inter.Block)
	return block
}

// SetBlock stores chain block index.
func (s *Store) SetBlockIndex(id hash.Event, n idx.Block) {
	if err := s.table.BlockHashes.Put(id.Bytes(), n.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetBlock returns stored block index.
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
