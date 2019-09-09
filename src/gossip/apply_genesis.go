package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

func (s *Store) ApplyGenesis(genesis *genesis.Genesis) (genesisAtropos hash.Event, genesisEvmState common.Hash, err error) {
	genesisHashFn := func(header *evm_core.EvmHeader) common.Hash {
		dummyAtropos := inter.NewEvent()
		// for nice-looking ID
		dummyAtropos.Epoch = 0
		dummyAtropos.Lamport = idx.Lamport(poset.EpochLen)
		// actual data hashed
		dummyAtropos.Extra = genesis.ExtraData
		dummyAtropos.ClaimedTime = header.Time
		dummyAtropos.TxHash = header.Root
		dummyAtropos.Creator = header.Coinbase

		return common.Hash(dummyAtropos.Hash())
	}

	evmBlock, err := evm_core.ApplyGenesis(s.table.Evm, genesis, genesisHashFn)

	block := inter.NewBlock(0, genesis.Time, hash.Events{hash.Event(evmBlock.Hash)}, hash.Event{})
	block.Root = evmBlock.Root
	block.Creator = evmBlock.Coinbase
	s.SetBlock(block)

	return block.Hash(), block.Root, err
}
