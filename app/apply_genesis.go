package app

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(net *lachesis.Config) (evmBlock *evmcore.EvmBlock, err error) {
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

	validatorsArr := []sfctype.SfcStakerAndID{}
	for _, validator := range net.Genesis.Alloc.Validators {
		staker := &sfctype.SfcStaker{
			Address:      validator.Address,
			CreatedEpoch: 0,
			CreatedTime:  net.Genesis.Time,
			StakeAmount:  validator.Stake,
			DelegatedMe:  big.NewInt(0),
		}
		s.SetSfcStaker(validator.ID, staker)
		validatorsArr = append(validatorsArr, sfctype.SfcStakerAndID{
			StakerID: validator.ID,
			Staker:   staker,
		})
	}
	s.SetEpochValidators(1, validatorsArr)

	return
}
