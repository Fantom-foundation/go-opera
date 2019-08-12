package posposet

import (
	"bytes"

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

func (s *Store) IsRoot(f idx.Frame, from hash.Peer, id hash.Event) bool {
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
	addrSize    = 32
	eventIdSize = 32
)

func (s *Store) ForEachRoot(f idx.Frame, do func(f idx.Frame, from hash.Peer, id hash.Event) bool) {
	err := s.epochTable.Roots.ForEachFrom(f.Bytes(), func(key, _ []byte) bool {
		if len(key) != frameSize+addrSize+eventIdSize {
			s.Fatalf("Roots table: Incorrect key len %d", len(key))
		}
		actualF := idx.BytesToFrame(key[:frameSize])
		actualCreator := hash.BytesToPeer(key[frameSize : frameSize+addrSize])
		actualId := hash.BytesToEvent(key[frameSize+addrSize:])
		if actualF < f {
			s.Fatalf("Roots table: frame %d < %d", actualF, f)
		}

		return do(actualF, actualCreator, actualId)
	})
	if err != nil {
		s.Fatal(err)
	}
}

func (s *Store) ForEachRootFrom(f idx.Frame, from hash.Peer, do func(f idx.Frame, from hash.Peer, id hash.Event) bool) {
	prefix := append(f.Bytes(), from.Bytes()...)

	err := s.epochTable.Roots.ForEach(prefix, func(key, _ []byte) bool {
		if len(key) != frameSize+addrSize+eventIdSize {
			s.Fatalf("Roots table: Incorrect key len %d", len(key))
		}
		actualF := idx.BytesToFrame(key[:frameSize])
		actualCreator := hash.BytesToPeer(key[frameSize : frameSize+addrSize])
		actualId := hash.BytesToEvent(key[frameSize+addrSize:])
		if actualF < f {
			s.Fatalf("Roots table: frame %d < %d", actualF, f)
		}
		if actualCreator != from {
			s.Fatalf("Roots table: creator %s != %s", actualCreator.String(), from.String())
		}

		return do(actualF, actualCreator, actualId)
	})
	if err != nil {
		s.Fatal(err)
	}
}
