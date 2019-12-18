package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
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
	s.SetEpochStats(0, &sfctype.EpochStats{
		Start:    net.Genesis.Time,
		End:      net.Genesis.Time,
		TotalFee: new(big.Int),
	})
	s.SetDirtyEpochStats(&sfctype.EpochStats{
		Start:    net.Genesis.Time,
		TotalFee: new(big.Int),
	})

	for _, validator := range net.Genesis.Alloc.GValidators {
		staker := &sfctype.SfcStaker{
			Address:      validator.Address,
			CreatedEpoch: 0,
			CreatedTime:  net.Genesis.Time,
			StakeAmount:  pos.StakeToBalance(validator.Stake),
			DelegatedMe:  big.NewInt(0),
		}
		s.SetSfcStaker(validator.ID, staker)
		s.SetEpochValidator(1, validator.ID, staker)
	}

	return genesisAtropos, genesisEvmState, nil
}
