package poset

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

func (s *Store) AddRoot(root *inter.Event) {
	key := bytes.Buffer{}
	key.Write(root.Frame.Bytes())
	key.Write(root.Creator.Bytes())
	key.Write(root.Hash().Bytes())

	if err := s.epochTable.Roots.Put(key.Bytes(), []byte{}); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) IsRoot(f idx.Frame, from common.Address, id hash.Event) bool {
	key := bytes.Buffer{}
	key.Write(f.Bytes())
	key.Write(from.Bytes())
	key.Write(id.Bytes())

	ok, err := s.epochTable.Roots.Has(key.Bytes())
	if err != nil {
		s.Fatal(err)
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
	for it.Next() {
		key := it.Key()
		if len(key) != frameSize+addrSize+eventIdSize {
			s.Fatalf("Roots table: Incorrect key len %d", len(key))
		}
		actualF := idx.BytesToFrame(key[:frameSize])
		actualCreator := common.BytesToAddress(key[frameSize : frameSize+addrSize])
		actualId := hash.BytesToEvent(key[frameSize+addrSize:])
		if actualF < f {
			s.Fatalf("Roots table: frame %d < %d", actualF, f)
		}

		if !do(actualF, actualCreator, actualId) {
			break
		}
	}
	if it.Error() != nil {
		s.Fatal(it.Error())
	}
	it.Release()
}

func (s *Store) ForEachRootFrom(f idx.Frame, from common.Address, do func(f idx.Frame, from common.Address, id hash.Event) bool) {
	prefix := append(f.Bytes(), from.Bytes()...)

	it := s.epochTable.Roots.NewIteratorWithPrefix(prefix)
	for it.Next() {
		key := it.Key()
		if len(key) != frameSize+addrSize+eventIdSize {
			s.Fatalf("Roots table: Incorrect key len %d", len(key))
		}
		actualF := idx.BytesToFrame(key[:frameSize])
		actualCreator := common.BytesToAddress(key[frameSize : frameSize+addrSize])
		actualId := hash.BytesToEvent(key[frameSize+addrSize:])
		if actualF < f {
			s.Fatalf("Roots table: frame %d < %d", actualF, f)
		}
		if actualCreator != from {
			s.Fatalf("Roots table: creator %s != %s", actualCreator.String(), from.String())
		}

		if !do(actualF, actualCreator, actualId) {
			break
		}
	}
	if it.Error() != nil {
		s.Fatal(it.Error())
	}
	it.Release()
}
