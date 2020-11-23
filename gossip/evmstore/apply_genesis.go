package evmstore

import (
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g opera.GenesisState) (evmBlock *evmcore.EvmBlock, err error) {
	return evmcore.ApplyGenesis(s.table.Evm, g)
}
