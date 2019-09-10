package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
)

func (s *Store) ApplyGenesis(genesis *genesis.Genesis) (genesisFiWitness hash.Event, genesisEvmState common.Hash, err error) {
	evmBlock, err := evm_core.ApplyGenesis(s.table.Evm, genesis)

	block := inter.NewBlock(0, genesis.Time, hash.Events{hash.Event(evmBlock.Hash)}, hash.Event{})
	block.Root = evmBlock.Root
	block.Creator = evmBlock.Coinbase
	s.SetBlock(block)

	return block.Hash(), block.Root, err
}
