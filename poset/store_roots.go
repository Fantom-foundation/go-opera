package poset

import (
	"bytes"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/poset/election"
)

func rootRecordBytes(r *election.RootAndSlot) []byte {
	key := bytes.Buffer{}
	key.Write(r.Slot.Frame.Bytes())
	key.Write(r.Slot.Validator.Bytes())
	key.Write(r.ID.Bytes())
	return key.Bytes()
}

// AddRoot stores the new root
// Not safe for concurrent use due to complex mutable cache!
func (s *Store) AddRoot(root *inter.Event) {
	r := election.RootAndSlot{
		Slot: election.Slot{
			Frame:     root.Frame,
			Validator: root.Creator,
		},
		ID: root.Hash(),
	}

	if err := s.epochTable.Roots.Put(rootRecordBytes(&r), []byte{}); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}

	// Add to cache.
	if s.cache.FrameRoots != nil {
		if c, ok := s.cache.FrameRoots.Get(root.Frame); ok {
			if rr, ok := c.([]election.RootAndSlot); ok {
				s.cache.FrameRoots.Add(root.Frame, append(rr, r))
			}
		}
	}
}

const (
	frameSize    = 4
	stakerIDSize = 4
	eventIDSize  = 32
)

// GetFrameRoots returns all the roots in the specified frame
// Not safe for concurrent use due to complex mutable cache!
func (s *Store) GetFrameRoots(f idx.Frame) []election.RootAndSlot {
	// Get data from LRU cache first.
	if s.cache.FrameRoots != nil {
		if c, ok := s.cache.FrameRoots.Get(f); ok {
			if rr, ok := c.([]election.RootAndSlot); ok {
				return rr
			}
		}
	}
	rr := make([]election.RootAndSlot, 0, 200)

	it := s.epochTable.Roots.NewIteratorWithPrefix(f.Bytes())
	defer it.Release()
	for it.Next() {
		key := it.Key()
		if len(key) != frameSize+stakerIDSize+eventIDSize {
			s.Log.Crit("Roots table: incorrect key len", "len", len(key))
		}
		r := election.RootAndSlot{
			Slot: election.Slot{
				Frame:     idx.BytesToFrame(key[:frameSize]),
				Validator: idx.BytesToStakerID(key[frameSize : frameSize+stakerIDSize]),
			},
			ID: hash.BytesToEvent(key[frameSize+stakerIDSize:]),
		}
		if r.Slot.Frame != f {
			s.Log.Crit("Roots table: invalid frame", "frame", r.Slot.Frame, "expected", f)
		}

		rr = append(rr, r)
	}
	if it.Error() != nil {
		s.Log.Crit("Failed to iterate keys", "err", it.Error())
	}

	// Add to cache.
	if s.cache.FrameRoots != nil {
		s.cache.FrameRoots.Add(f, rr)
	}

	return rr
}
