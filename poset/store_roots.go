package poset

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func (s *Store) AddRoot(root *inter.Event) {
	key := bytes.Buffer{}
	key.Write(root.Frame.Bytes())
	key.Write(root.Creator.Bytes())
	key.Write(root.Hash().Bytes())

	if err := s.epochTable.Roots.Put(key.Bytes(), []byte{}); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) IsRoot(f idx.Frame, from common.Address, id hash.Event) bool {
	key := bytes.Buffer{}
	key.Write(f.Bytes())
	key.Write(from.Bytes())
	key.Write(id.Bytes())

	ok, err := s.epochTable.Roots.Has(key.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	return ok
}

const (
	frameSize   = 4
	addrSize    = 20
	eventIdSize = 32
)

func (s *Store) ForEachRoot(f idx.Frame, do func(f idx.Frame, from common.Address, root hash.Event) bool) {
	it := s.epochTable.Roots.NewIteratorWithStart(f.Bytes())
	defer it.Release()
	for it.Next() {
		key := it.Key()
		if len(key) != frameSize+addrSize+eventIdSize {
			s.Log.Crit("Roots table: incorrect key len", "len", len(key))
		}
		actualF := idx.BytesToFrame(key[:frameSize])
		actualCreator := common.BytesToAddress(key[frameSize : frameSize+addrSize])
		actualID := hash.BytesToEvent(key[frameSize+addrSize:])
		if actualF < f {
			s.Log.Crit("Roots table: invalid frame", "frame", f, "expected", actualF)
		}

		if !do(actualF, actualCreator, actualID) {
			break
		}
	}
	if it.Error() != nil {
		s.Log.Crit("Failed to iterate keys", "err", it.Error())
	}
}

func (s *Store) ForEachRootFrom(f idx.Frame, from common.Address, do func(f idx.Frame, from common.Address, id hash.Event) bool) {
	prefix := append(f.Bytes(), from.Bytes()...)

	it := s.epochTable.Roots.NewIteratorWithPrefix(prefix)
	defer it.Release()
	for it.Next() {
		key := it.Key()
		if len(key) != frameSize+addrSize+eventIdSize {
			s.Log.Crit("Roots table: incorrect key len", "len", len(key))
		}
		actualF := idx.BytesToFrame(key[:frameSize])
		actualCreator := common.BytesToAddress(key[frameSize : frameSize+addrSize])
		actualID := hash.BytesToEvent(key[frameSize+addrSize:])
		if actualF < f {
			s.Log.Crit("Roots table: invalid frame", "frame", f, "expected", actualF)
		}
		if actualCreator != from {
			s.Log.Crit("Roots table: invalid creator", "creator", from.String(), "expected", actualCreator.String())
		}

		if !do(actualF, actualCreator, actualID) {
			break
		}
	}
	if it.Error() != nil {
		s.Log.Crit("Failed to iterate keys", "err", it.Error())
	}
}
