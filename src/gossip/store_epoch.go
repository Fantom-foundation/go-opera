package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
)

type (
	epochStore struct {
		Headers kvdb.Database `table:"header_"`
		Tips    kvdb.Database `table:"tips_"`
		Heads   kvdb.Database `table:"heads_"`
	}
)

// getEpochStore is not safe for concurrent use.
func (s *Store) getEpochStore(epoch idx.SuperFrame) *epochStore {
	tables := s.getTmpDb("epoch", uint64(epoch), func(db kvdb.Database) interface{} {
		es := &epochStore{}
		kvdb.MigrateTables(es, db)
		return es
	})
	if tables == nil {
		return nil
	}

	return tables.(*epochStore)
}

// delEpochStore is not safe for concurrent use.
func (s *Store) delEpochStore(epoch idx.SuperFrame) {
	s.delTmpDb("epoch", uint64(epoch))
}

func (s *Store) SetLastEvent(epoch idx.SuperFrame, from hash.Peer, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := from.Bytes()
	if err := es.Tips.Put(key, id.Bytes()); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) GetLastEvent(epoch idx.SuperFrame, from hash.Peer) *hash.Event {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	key := from.Bytes()
	idBytes, err := es.Tips.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if idBytes == nil {
		return nil
	}
	id := hash.BytesToEvent(idBytes)
	return &id
}

// SetEventHeader returns stored event header.
func (s *Store) SetEventHeader(epoch idx.SuperFrame, h hash.Event, e *inter.EventHeaderData) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := h.Bytes()

	s.set(es.Headers, key, e)
}

// GetEventHeader returns stored event header.
func (s *Store) GetEventHeader(epoch idx.SuperFrame, h hash.Event) *inter.EventHeaderData {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	key := h.Bytes()

	w, _ := s.get(es.Headers, key, &inter.EventHeaderData{}).(*inter.EventHeaderData)
	return w
}

// DelEventHeader removes stored event header.
func (s *Store) DelEventHeader(epoch idx.SuperFrame, h hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := h.Bytes()
	err := es.Headers.Delete(key)
	if err != nil {
		s.Fatal(err)
	}
}
