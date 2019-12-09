package gaspowercheck

import (
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

var (
	// ErrWrongGasPowerLeft indicates that event's GasPowerLeft is miscalculated.
	ErrWrongGasPowerLeft = errors.New("event has wrong GasPowerLeft")
)

// DagReader is accessed by the validator to get the current state.
type DagReader interface {
	GetEpochValidators() (*pos.Validators, idx.Epoch)
	GetPrevEpochLastHeaders() (inter.HeadersByCreator, idx.Epoch)
	GetPrevEpochEndTime() (inter.Timestamp, idx.Epoch)
}

// Checker which checks gas power
type Checker struct {
	config *lachesis.GasPowerConfig
	reader DagReader
}

// New Checker for gas power
func New(config *lachesis.GasPowerConfig, reader DagReader) *Checker {
	return &Checker{
		config: config,
		reader: reader,
	}
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
	validators *pos.Validators,
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
func CalcGasPower(
	e *inter.EventHeaderData,
	selfParent *inter.EventHeaderData,
	validators *pos.Validators,
	lastHeaders inter.HeadersByCreator,
	prevEpochEnd inter.Timestamp,
	config *lachesis.GasPowerConfig,
) uint64 {
	gasPowerPerH, maxGasPower, startup := calcValidatorGasPowerPerH(e.Creator, validators, config)

	var prevGasPowerLeft uint64
	var prevMedianTime inter.Timestamp

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
		prevMedianTime = prevEpochEnd
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

// Validate event
func (v *Checker) Validate(e *inter.Event, selfParent *inter.EventHeaderData) error {
	validators, epoch := v.reader.GetEpochValidators()
	lastHeaders, prevEpoch1 := v.reader.GetPrevEpochLastHeaders()
	epochEndTime, prevEpoch2 := v.reader.GetPrevEpochEndTime()
	// check that all the data is for the same epoch
	if epoch != e.Epoch {
		return epochcheck.ErrNotRelevant
	}
	if prevEpoch1+1 != e.Epoch {
		return epochcheck.ErrNotRelevant
	}
	if prevEpoch2+1 != e.Epoch {
		return epochcheck.ErrNotRelevant
	}

	gasPower := CalcGasPower(&e.EventHeaderData, selfParent, validators, lastHeaders, epochEndTime, v.config)
	if e.GasPowerLeft+e.GasPowerUsed != gasPower { // GasPowerUsed is checked in basic_check
		return ErrWrongGasPowerLeft
	}
	return nil
}
