package evmstore

import (
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(net *opera.Config) (evmBlock *evmcore.EvmBlock, err error) {
	return evmcore.ApplyGenesis(s.table.Evm, net)
}
