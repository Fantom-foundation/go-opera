package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) *state.DB {
	db, err := state.New(from, s.table.Balances)
	if err != nil {
		s.Fatal(err)
	}
	return db
}

// SetCheckpoint save checkpoint.
// Checkpoint is seldom read; so no cache.
func (s *Store) SetCheckpoint(cp *checkpoint) {
	const key = "current"
	s.set(s.table.Checkpoint, []byte(key), cp)

}

// GetCheckpoint returns stored checkpoint.
// State is seldom read; so no cache.
func (s *Store) GetCheckpoint() *checkpoint {
	const key = "current"
	w, _ := s.get(s.table.Checkpoint, []byte(key), &checkpoint{}).(*checkpoint)
	return w
}
