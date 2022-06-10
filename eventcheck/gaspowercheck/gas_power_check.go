package gaspowercheck

import (
	"errors"
	"math/big"
	"time"

	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
)

var (
	// ErrWrongGasPowerLeft indicates that event's GasPowerLeft is miscalculated.
	ErrWrongGasPowerLeft = errors.New("event has wrong GasPowerLeft")
)

type ValidatorState struct {
	PrevEpochEvent iblockproc.EventInfo
	GasRefund      uint64
}

// ValidationContext for gaspower checking
type ValidationContext struct {
	Epoch           idx.Epoch
	Configs         [inter.GasPowerConfigs]Config
	EpochStart      inter.Timestamp
	Validators      *pos.Validators
	ValidatorStates []ValidatorState
}

// Reader is accessed by the validator to get the current state.
type Reader interface {
	GetValidationContext() *ValidationContext
}

// Config for gaspower checking. There'll be 2 different configs for short-term and long-term gas power checks.
type Config struct {
	Idx                int
	AllocPerSec        uint64
	MaxAllocPeriod     inter.Timestamp
	MinEnsuredAlloc    uint64
	StartupAllocPeriod inter.Timestamp
	MinStartupGas      uint64
}

// Checker which checks gas power
type Checker struct {
	reader Reader
}

// New Checker for gas power
func New(reader Reader) *Checker {
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

// CalcGasPower calculates available gas power for the event, i.e. how many gas its content may consume
func (v *Checker) CalcGasPower(e inter.EventI, selfParent inter.EventI) (inter.GasPowerLeft, error) {
	ctx := v.reader.GetValidationContext()
	// check that all the data is for the same epoch
	if ctx.Epoch != e.Epoch() {
		return inter.GasPowerLeft{}, epochcheck.ErrNotRelevant
	}

	var res inter.GasPowerLeft
	for i := range ctx.Configs {
		res.Gas[i] = calcGasPower(e, selfParent, ctx, ctx.Configs[i])
	}

	return res, nil
}

func calcGasPower(e inter.EventI, selfParent inter.EventI, ctx *ValidationContext, config Config) uint64 {
	var prevGasPowerLeft uint64
	var prevTime inter.Timestamp

	if e.SelfParent() != nil {
		prevGasPowerLeft = selfParent.GasPowerLeft().Gas[config.Idx]
		prevTime = selfParent.MedianTime()
	} else {
		validatorState := ctx.ValidatorStates[ctx.Validators.GetIdx(e.Creator())]
		if validatorState.PrevEpochEvent.ID != hash.ZeroEvent {
			prevGasPowerLeft = validatorState.PrevEpochEvent.GasPowerLeft.Gas[config.Idx]
			prevTime = validatorState.PrevEpochEvent.Time
		} else {
			prevGasPowerLeft = 0
			prevTime = ctx.EpochStart
		}
		prevGasPowerLeft += validatorState.GasRefund
	}

	return CalcValidatorGasPower(e, e.MedianTime(), prevTime, prevGasPowerLeft, ctx.Validators, config)
}

func CalcValidatorGasPower(e inter.EventI, eTime, prevTime inter.Timestamp, prevGasPowerLeft uint64, validators *pos.Validators, config Config) uint64 {
	gasPowerPerSec, maxGasPower, startup := CalcValidatorGasPowerPerSec(e.Creator(), validators, config)

	if e.SelfParent() == nil {
		if prevGasPowerLeft < startup {
			prevGasPowerLeft = startup
		}
	}

	if prevTime > eTime {
		prevTime = eTime
	}

	gasPowerAllocatedBn := new(big.Int).SetUint64(uint64(eTime - prevTime))
	mul(gasPowerAllocatedBn, gasPowerPerSec)
	div(gasPowerAllocatedBn, uint64(time.Second))

	gasPower := gasPowerAllocatedBn.Uint64() + prevGasPowerLeft
	if gasPower > maxGasPower {
		gasPower = maxGasPower
	}

	return gasPower
}

func CalcValidatorGasPowerPerSec(
	validator idx.ValidatorID,
	validators *pos.Validators,
	config Config,
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
	div(validatorGasPowerPerSecBn, uint64(validators.TotalWeight()))
	perSec = validatorGasPowerPerSecBn.Uint64()

	maxGasPower = perSec * (uint64(gas.MaxAllocPeriod) / uint64(time.Second))
	if maxGasPower < gas.MinEnsuredAlloc {
		maxGasPower = gas.MinEnsuredAlloc
	}

	startup = perSec * (uint64(gas.StartupAllocPeriod) / uint64(time.Second))
	if startup < gas.MinStartupGas {
		startup = gas.MinStartupGas
	}

	return
}

// Validate event
func (v *Checker) Validate(e inter.EventI, selfParent inter.EventI) error {
	gasPowers, err := v.CalcGasPower(e, selfParent)
	if err != nil {
		return err
	}
	for i := range gasPowers.Gas {
		if e.GasPowerLeft().Gas[i]+e.GasPowerUsed() != gasPowers.Gas[i] { // GasPowerUsed is checked in basic_check
			return ErrWrongGasPowerLeft
		}
	}
	return nil
}
