package gossip

import (
	"bytes"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

const (
	epochSize   = 4
	packSize    = 4
	eventIdSize = 32
)

func (s *Store) GetPackInfo(epoch idx.SuperFrame, idx idx.Pack) PackInfo {
	key := bytes.Buffer{}
	key.Write(epoch.Bytes())
	key.Write(idx.Bytes())

	w, exists := s.get(s.table.PackInfos, key.Bytes(), &PackInfo{}).(*PackInfo)
	if !exists {
		return PackInfo{
			Index: idx,
		}
	}
	return *w
}

func (s *Store) SetPackInfo(epoch idx.SuperFrame, idx idx.Pack, value PackInfo) {
	key := bytes.Buffer{}
	key.Write(epoch.Bytes())
	key.Write(idx.Bytes())

	s.set(s.table.PackInfos, key.Bytes(), value)
}

func (s *Store) AddToPack(epoch idx.SuperFrame, idx idx.Pack, e hash.Event) {
	key := bytes.Buffer{}
	key.Write(epoch.Bytes())
	key.Write(idx.Bytes())
	key.Write(e.Bytes())

	err := s.table.Packs.Put(key.Bytes(), []byte{})
	if err != nil {
		s.Fatal(err)
	}
}

func (s *Store) GetPack(epoch idx.SuperFrame, idx idx.Pack) hash.Events {
	prefix := bytes.Buffer{}
	prefix.Write(epoch.Bytes())
	prefix.Write(idx.Bytes())

	res := make(hash.Events, 0, hardLimitItems)

	err := s.table.Packs.ForEach(prefix.Bytes(), func(key, _ []byte) bool {
		if len(key) != epochSize+packSize+eventIdSize {
			s.Fatalf("packs table: Incorrect key len %d", len(key))
		}
		res.Add(hash.BytesToEvent(key[epochSize+packSize:]))

		return true
	})
	if err != nil {
		s.Fatal(err)
	}
	return res
}

func (s *Store) GetPacksNum(epoch idx.SuperFrame) idx.Pack {
	b, err := s.table.PacksNum.Get(epoch.Bytes())
	if err != nil {
		s.Fatal(err)
	}
	if b == nil {
		return idx.Pack(1)
	}
	return idx.BytesToPack(b)
}

func (s *Store) SetPacksNum(epoch idx.SuperFrame, num idx.Pack) {
	err := s.table.PacksNum.Put(epoch.Bytes(), num.Bytes())
	if err != nil {
		s.Fatal(err)
	}
}
