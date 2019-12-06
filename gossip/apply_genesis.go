package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

func (s *Store) ApplyGenesis(net *lachesis.Config) (genesisAtropos hash.Event, genesisEvmState common.Hash, err error) {
	evmBlock, err := evmcore.ApplyGenesis(s.table.Evm, net)
	if err != nil {
		return hash.Event{}, common.Hash{}, err
	}

	genesisAtropos = hash.Event(evmBlock.Hash)
	genesisEvmState = evmBlock.Root

	if s.GetBlock(0) != nil {
		return genesisAtropos, genesisEvmState, nil
	}
	// first time genesis is applied

	block := inter.NewBlock(0,
		net.Genesis.Time,
		genesisAtropos,
		hash.Event{},
		hash.Events{genesisAtropos},
	)

	block.Root = evmBlock.Root
	s.SetBlock(block)
	s.SetEpochStats(0, &EpochStats{
		Start:    net.Genesis.Time,
		End:      net.Genesis.Time,
		TotalFee: new(big.Int),
	})
	s.SetDirtyEpochStats(&EpochStats{
		Start:    net.Genesis.Time,
		TotalFee: new(big.Int),
	})

	for i, validator := range net.Genesis.Validators.SortedAddresses() { // sort validators to get deterministic stakerIDs
		stakerID := uint64(i + 1)
		stakeAmount := net.Genesis.Validators.Get(validator)

		staker := &SfcStaker{
			Address:      validator,
			CreatedEpoch: 0,
			CreatedTime:  net.Genesis.Time,
			StakeAmount:  pos.StakeToBalance(stakeAmount),
			DelegatedMe:  big.NewInt(0),
		}
		s.SetSfcStaker(stakerID, staker)
		s.SetEpochValidator(0, stakerID, staker)
	}

	return genesisAtropos, genesisEvmState, nil
}
