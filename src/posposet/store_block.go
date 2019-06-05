package posposet

import "github.com/Fantom-foundation/go-lachesis/src/posposet/wire"

// SetBlock stores chain block.
// State is seldom read; so no cache.
func (s *Store) SetBlock(b *Block) {
	s.set(s.Blocks, intToBytes(b.Index), b.ToWire())
}

// GetBlock returns stored block.
// State is seldom read; so no cache.
func (s *Store) GetBlock(n uint64) *Block {
	w, _ := s.get(s.Blocks, intToBytes(n), &wire.Block{}).(*wire.Block)
	return WireToBlock(w)
}
