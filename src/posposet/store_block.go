package posposet

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

// SetEventsBlockNum stores num of block includes events.
func (s *Store) SetEventsBlockNum(num idx.Block, ee ...*inter.Event) {
	val := num.Bytes()

	for _, e := range ee {
		key := e.Hash()

		if err := s.table.Event2Block.Put(key.Bytes(), val); err != nil {
			s.Fatal(err)
		}
	}
}

// GetEventBlockNum returns num of block includes event.
func (s *Store) GetEventBlockNum(e hash.Event) *idx.Block {
	key := e.Bytes()
	buf, err := s.table.Event2Block.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	val := idx.BytesToBlock(buf)
	return &val
}
