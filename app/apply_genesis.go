package app

import (
	"math/big"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(net *opera.Config) (evmBlock *evmcore.EvmBlock, err error) {
	evmBlock, err = evmcore.ApplyGenesis(s.table.Evm, net)
	if err != nil {
		return
	}

	// calc total pre-minted supply
	totalSupply := big.NewInt(0)
	for _, account := range net.Genesis.Alloc.Accounts {
		totalSupply.Add(totalSupply, account.Balance)
	}
	s.SetTotalSupply(totalSupply)

	return
}
