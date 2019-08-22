package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// ApplyBlock execs ordered txns on state.
// TODO: replace with EVM transactions
func (s *Service) ApplyBlock(block *inter.Block, stateHash hash.Hash, members pos.Members) (hash.Hash, pos.Members) {
	statedb := s.store.StateDB(stateHash)
	newStateHash, err := statedb.Commit(true)
	if err != nil {
		panic(err)
	}

	s.store.SetBlock(block)

	return hash.Hash(newStateHash), members
}
