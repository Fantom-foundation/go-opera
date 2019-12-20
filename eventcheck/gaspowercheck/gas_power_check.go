package gaspowercheck

import (
	"errors"
	"math/big"
	"time"

	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

var (
	// ErrWrongGasPowerLeft indicates that event's GasPowerLeft is miscalculated.
	ErrWrongGasPowerLeft = errors.New("event has wrong GasPowerLeft")
)

// ValidationContext for gaspower checking
type ValidationContext struct {
	Epoch                idx.Epoch
	Configs              [2]Config
	Validators           *pos.Validators
	PrevEpochLastHeaders inter.HeadersByCreator
	PrevEpochEndTime     inter.Timestamp
	PrevEpochRefunds     map[idx.StakerID]uint64
}

// DagReader is accessed by the validator to get the current state.
type DagReader interface {
	GetValidationContext() *ValidationContext
}

// Config for gaspower checking. There'll be 2 different configs for short-term and long-term gas power checks.
type Config struct {
	Idx                int
	AllocPerSec        uint64
	MaxAllocPeriod     inter.Timestamp
	StartupAllocPeriod inter.Timestamp
	MinStartupGas      uint64
}

// Checker which checks gas power
type Checker struct {
	reader DagReader
}

// New Checker for gas power
func New(reader DagReader) *Checker {
	return &Checker{
		reader: reader,
	}
}

func mul(a *big.Int, b uint64) {
	a.Mul(a, new(big.Int).SetUint64(b))
}

func div(a *big.Int, b uint64) {
	a.Div(a, new(big.Int).SetUint64(b))
}

func calcValidatorGasPowerPerSec(
	validator idx.StakerID,
	validators *pos.Validators,
	config *Config,
) (
	perSec uint64,
	maxGasPower uint64,
	startup uint64,
) {
	stake := validators.Get(validator)
	if stake == 0 {
		return 0, 0, 0
	}

	gas := config

	validatorGasPowerPerSecBn := new(big.Int).SetUint64(gas.AllocPerSec)
	mul(validatorGasPowerPerSecBn, uint64(stake))
	div(validatorGasPowerPerSecBn, uint64(validators.TotalStake()))
	perSec = validatorGasPowerPerSecBn.Uint64()

	maxGasPower = perSec * (uint64(gas.MaxAllocPeriod) / uint64(time.Second))

	startup = perSec * (uint64(gas.StartupAllocPeriod) / uint64(time.Second))
	if startup < gas.MinStartupGas {
		startup = gas.MinStartupGas
	}

	return
}

// CalcGasPower calculates available gas power for the event, i.e. how many gas its content may consume
func (v *Checker) CalcGasPower(e *inter.EventHeaderData, selfParent *inter.EventHeaderData) (inter.GasPowerLeft, error) {
	ctx := v.reader.GetValidationContext()
	// check that all the data is for the same epoch
	if ctx.Epoch != e.Epoch {
		return inter.GasPowerLeft{}, epochcheck.ErrNotRelevant
	}

	var res inter.GasPowerLeft
	for i := range ctx.Configs {
		res.Gas[i] = calcGasPower(e, selfParent, ctx, &ctx.Configs[i])
	}

	return res, nil
}

func calcGasPower(e *inter.EventHeaderData, selfParent *inter.EventHeaderData, ctx *ValidationContext, config *Config) uint64 {
	gasPowerPerSec, maxGasPower, startup := calcValidatorGasPowerPerSec(e.Creator, ctx.Validators, config)

	var prevGasPowerLeft uint64
	var prevMedianTime inter.Timestamp

	if e.SelfParent() != nil {
		prevGasPowerLeft = selfParent.GasPowerLeft.Gas[config.Idx]
		prevMedianTime = selfParent.MedianTime
	} else if prevConfirmedHeader := ctx.PrevEpochLastHeaders[e.Creator]; prevConfirmedHeader != nil {
		prevGasPowerLeft = prevConfirmedHeader.GasPowerLeft.Gas[config.Idx] + ctx.PrevEpochRefunds[e.Creator]
		if prevGasPowerLeft < startup {
			prevGasPowerLeft = startup
		}
		prevMedianTime = prevConfirmedHeader.MedianTime
	} else {
		prevGasPowerLeft = startup + ctx.PrevEpochRefunds[e.Creator]
		prevMedianTime = ctx.PrevEpochEndTime
	}

	if prevMedianTime > e.MedianTime {
		prevMedianTime = e.MedianTime // do not change e.MedianTime
	}

	gasPowerAllocatedBn := new(big.Int).SetUint64(uint64(e.MedianTime - prevMedianTime))
	mul(gasPowerAllocatedBn, gasPowerPerSec)
	div(gasPowerAllocatedBn, uint64(time.Second))

	gasPower := gasPowerAllocatedBn.Uint64() + prevGasPowerLeft
	if gasPower > maxGasPower {
		gasPower = maxGasPower
	}

	return gasPower
}

// Validate event
func (v *Checker) Validate(e *inter.Event, selfParent *inter.EventHeaderData) error {
	gasPowers, err := v.CalcGasPower(&e.EventHeaderData, selfParent)
	if err != nil {
		return err
	}
	for i := range gasPowers.Gas {
		if e.GasPowerLeft.Gas[i]+e.GasPowerUsed != gasPowers.Gas[i] { // GasPowerUsed is checked in basic_check
			return ErrWrongGasPowerLeft
		}
	}
	return nil
}
