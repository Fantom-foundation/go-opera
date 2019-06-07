package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// SetBlock stores chain block.
func (s *Store) SetBlock(b *inter.Block) {
	key := common.IntToBytes(b.Index)
	s.set(s.table.Blocks, key, b.ToWire())
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n uint64) *inter.Block {
	key := common.IntToBytes(n)
	w, _ := s.get(s.table.Blocks, key, &wire.Block{}).(*wire.Block)
	return inter.WireToBlock(w)
}

// SetEventsBlockNum stores num of block includes events.
// TODO: use batch
func (s *Store) SetEventsBlockNum(num uint64, ee ...*inter.Event) {
	val := common.IntToBytes(num)

	for _, e := range ee {
		key := e.Hash()

		if err := s.table.Event2Block.Put(key.Bytes(), val); err != nil {
			s.Fatal(err)
		}

		if s.cache.Event2Block != nil {
			s.cache.Event2Block.Add(key, num)
		}
	}
}

// GetEventBlockNum returns num of block includes event.
func (s *Store) GetEventBlockNum(e hash.Event) *uint64 {
	if s.cache.Event2Block != nil {
		if n, ok := s.cache.Event2Block.Get(e); ok {
			num := n.(uint64)
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

	val := common.BytesToInt(buf)
	return &val
}
