package gossip

import (
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
