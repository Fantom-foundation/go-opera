package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

func (s *Store) ApplyGenesis(net *lachesis.Config) (genesisAtropos hash.Event, genesisEvmState common.Hash, err error) {
	evmBlock, err := evmcore.ApplyGenesis(s.table.Evm, net)
	if err != nil {
		return
	}

	block := inter.NewBlock(0,
		net.Genesis.Time,
		hash.Event(evmBlock.Hash),
		hash.Event{},
		hash.Events{hash.Event(evmBlock.Hash)},
	)

	block.Root = evmBlock.Root
	s.SetBlock(block)
	genesisAtropos = block.Hash()
	genesisEvmState = block.Root

	return
}
