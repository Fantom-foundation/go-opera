package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// ApplyBlock execs ordered txns on state.
// TODO: replace with EVM transactions
func (s *Service) ApplyBlock(block *inter.Block, stateHash common.Hash, members pos.Members) (common.Hash, pos.Members) {
	statedb := s.store.StateDB(stateHash)
	newStateHash, err := statedb.Commit(true)
	if err != nil {
		panic(err)
	}

	s.store.SetBlock(block)

	return newStateHash, members
}
