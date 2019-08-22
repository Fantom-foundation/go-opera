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

func (s *Store) ApplyGenesis(genesis *genesis.Genesis) (genesisFiWitness hash.Event, genesisEvmState hash.Hash, err error) {
	genesisHashFn := func(header *evm_core.EvmHeader) common.Hash {
		dummyFiWitness := inter.NewEvent()
		// for nice-looking ID
		dummyFiWitness.Epoch = 0
		dummyFiWitness.Lamport = idx.Lamport(poset.SuperFrameLen)
		// actual data hashed
		dummyFiWitness.Extra = genesis.ExtraData
		dummyFiWitness.ClaimedTime = header.Time
		dummyFiWitness.TxHash = hash.Hash(header.Root)
		//dummyFiWitness.Creator = header.Coinbase TODO
		return common.Hash(dummyFiWitness.Hash())
	}

	genesisHash, rootHash, err := evm_core.ApplyGenesis(s.table.Evm, genesis, genesisHashFn)
	return hash.Event(genesisHash), hash.Hash(rootHash), err
}
