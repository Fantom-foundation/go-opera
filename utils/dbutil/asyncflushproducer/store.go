package asyncflushproducer

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
)

type store struct {
	kvdb.Store
	CloseFn func() error
}

func (s *store) Close() error {
	return s.CloseFn()
}
