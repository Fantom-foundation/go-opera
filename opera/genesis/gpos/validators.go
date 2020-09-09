package gpos

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
)

type (
	// Validator is a helper structure to define genesis validators
	Validator struct {
		ID      idx.ValidatorID
		Address common.Address
		Stake   *big.Int
	}

	Validators []Validator
)

// Build converts Validators to Validators
func (gv Validators) Build() *pos.Validators {
	builder := pos.NewBigBuilder()
	for _, validator := range gv {
		builder.Set(validator.ID, validator.Stake)
	}
	return builder.Build()
}

// TotalStake returns sum of stakes
func (gv Validators) TotalStake() *big.Int {
	totalStake := new(big.Int)
	for _, validator := range gv {
		totalStake.Add(totalStake, validator.Stake)
	}
	return totalStake
}

// Map converts Validators to map
func (gv Validators) Map() map[idx.ValidatorID]Validator {
	validators := map[idx.ValidatorID]Validator{}
	for _, validator := range gv {
		validators[validator.ID] = validator
	}
	return validators
}

// Addresses returns not sorted genesis addresses
func (gv Validators) Addresses() []common.Address {
	res := make([]common.Address, 0, len(gv))
	for _, v := range gv {
		res = append(res, v.Address)
	}
	return res
}
