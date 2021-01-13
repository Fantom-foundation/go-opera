package evmstore

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g opera.Genesis) (evmBlock *evmcore.EvmBlock, err error) {
	// state
	statedb, err := s.StateDB(hash.Hash{})
	if err != nil {
		return nil, err
	}
	return evmcore.ApplyGenesis(statedb, g, 32*opt.MiB)
}
