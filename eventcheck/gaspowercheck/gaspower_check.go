package gaspowercheck

import (
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

// DagReader is accessed by the validator to get the current state.
type DagReader interface {
	GetValidators() pos.Validators
	GetLastHeaders() inter.HeadersByCreator
	GetPrevEpochTime() inter.Timestamp
}

// Checker which require only parents list + current epoch info
type Checker struct {
	config *lachesis.GasPowerConfig
	reader DagReader
}

func mul(a *big.Int, b uint64) {
	a.Mul(a, new(big.Int).SetUint64(b))
}

func div(a *big.Int, b uint64) {
	a.Div(a, new(big.Int).SetUint64(b))
}

func sub(a *big.Int, b uint64) {
	a.Sub(a, new(big.Int).SetUint64(b))
}

func calcValidatorGasPowerPerH(
	validator common.Address,
	validators pos.Validators,
	config *lachesis.GasPowerConfig,
) (
	perHour uint64,
	maxGasPower uint64,
	startup uint64,
) {
	stake := validators.Get(validator)
	if stake == 0 {
		return 0, 0, 0
	}

	gas := config

	validatorGasPowerPerHBn := new(big.Int).SetUint64(gas.TotalPerH)
	mul(validatorGasPowerPerHBn, uint64(stake))
	div(validatorGasPowerPerHBn, uint64(validators.TotalStake()))
	perHour = validatorGasPowerPerHBn.Uint64()

	validatorMaxGasPowerBn := new(big.Int).Set(validatorGasPowerPerHBn)
	mul(validatorMaxGasPowerBn, uint64(gas.MaxGasPowerPeriod))
	div(validatorMaxGasPowerBn, uint64(time.Hour))
	maxGasPower = validatorMaxGasPowerBn.Uint64()

	validatorStartupBn := new(big.Int).Set(validatorGasPowerPerHBn)
	mul(validatorStartupBn, uint64(gas.StartupPeriod))
	div(validatorStartupBn, uint64(time.Hour))
	startup = validatorStartupBn.Uint64()
	if startup < gas.MinStartupGasPower {
		startup = gas.MinStartupGasPower
	}

	return
}

// CalcGasPower calculates available gas power for the event, i.e. how many gas its content may consume
func (v *Checker) CalcGasPower(
	e *inter.EventHeaderData,
	selfParent *inter.EventHeaderData,
) uint64 {
	validators := v.reader.GetValidators()

	gasPowerPerH, maxGasPower, startup := calcValidatorGasPowerPerH(e.Creator, validators, v.config)

	var prevGasPowerLeft uint64
	var prevMedianTime inter.Timestamp

	lastHeaders := v.reader.GetLastHeaders()

	if e.SelfParent() != nil {
		prevGasPowerLeft = selfParent.GasPowerLeft
		prevMedianTime = selfParent.MedianTime
	} else if prevConfirmedHeader := lastHeaders[e.Creator]; prevConfirmedHeader != nil {
		prevGasPowerLeft = prevConfirmedHeader.GasPowerLeft
		if prevGasPowerLeft < startup {
			prevGasPowerLeft = startup
		}
		prevMedianTime = prevConfirmedHeader.MedianTime
	} else {
		prevGasPowerLeft = startup
		prevMedianTime = v.reader.GetPrevEpochTime()
	}

	if prevMedianTime > e.MedianTime {
		prevMedianTime = e.MedianTime // do not change e.MedianTime
	}

	gasPowerAllocatedBn := new(big.Int).SetUint64(uint64(e.MedianTime))
	sub(gasPowerAllocatedBn, uint64(prevMedianTime))
	mul(gasPowerAllocatedBn, gasPowerPerH)
	div(gasPowerAllocatedBn, uint64(time.Hour))

	gasPower := gasPowerAllocatedBn.Uint64() + prevGasPowerLeft
	if gasPower > maxGasPower {
		gasPower = maxGasPower
	}

	return gasPower
}
