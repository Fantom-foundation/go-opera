package switchable

import (
	"sync"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/synced"
)

type Snapshot struct {
	kvdb.Snapshot
	mu sync.RWMutex
}

func (s *Snapshot) SwitchTo(snap kvdb.Snapshot) kvdb.Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.Snapshot
	s.Snapshot = synced.WrapSnapshot(snap, &s.mu)
	return old
}

func Wrap(snap kvdb.Snapshot) *Snapshot {
	s := &Snapshot{}
	s.SwitchTo(snap)
	return s
}
