package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// SetBlock stores chain block.
func (s *Store) SetBlock(b *inter.Block) {
	s.set(s.table.Blocks, b.Index.Bytes(), b.ToWire())
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n idx.Block) *inter.Block {
	w, _ := s.get(s.table.Blocks, n.Bytes(), &wire.Block{}).(*wire.Block)
	return inter.WireToBlock(w)
}

// SetEventsBlockNum stores num of block includes events.
// TODO: use batch
func (s *Store) SetEventsBlockNum(num idx.Block, ee ...*inter.Event) {
	val := num.Bytes()

	for _, e := range ee {
		key := e.Hash()

		if err := s.table.Event2Block.Put(key.Bytes(), val); err != nil {
			s.Fatal(err)
		}

		if s.cache.Event2Block != nil {
			s.cache.Event2Block.Add(key, num)

			event2BlockCacheCap.Update(int64(
				s.cache.Event2Block.Len()))
		}
	}
}

// GetEventBlockNum returns num of block includes event.
func (s *Store) GetEventBlockNum(e hash.Event) *idx.Block {
	if s.cache.Event2Block != nil {
		if n, ok := s.cache.Event2Block.Get(e); ok {
			num := n.(idx.Block)
			return &num
		}
	}

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
