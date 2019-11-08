package poset

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

func mul(a *big.Int, b uint64) {
	a.Mul(a, new(big.Int).SetUint64(b))
}

func div(a *big.Int, b uint64) {
	a.Div(a, new(big.Int).SetUint64(b))
}

func sub(a *big.Int, b uint64) {
	a.Sub(a, new(big.Int).SetUint64(b))
}

func (p *Poset) calcValidatorGasPowerPerH(validator common.Address) (perHour, maxGasPower, startup uint64) {
	stake, ok := p.Validators[validator]
	if !ok {
		return 0, 0, 0
	}

	gas := p.dag.GasPower

	validatorGasPowerPerHBn := new(big.Int).SetUint64(gas.TotalPerH)
	mul(validatorGasPowerPerHBn, uint64(stake))
	div(validatorGasPowerPerHBn, uint64(p.Validators.TotalStake()))
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
func (p *Poset) CalcGasPower(e *inter.EventHeaderData) uint64 {
	gasPowerPerH, maxGasPower, startup := p.calcValidatorGasPowerPerH(e.Creator)

	var prevGasPowerLeft uint64
	var prevMedianTime inter.Timestamp

	if e.SelfParent() != nil {
		selfParent := p.GetEventHeader(e.Epoch, *e.SelfParent())
		prevGasPowerLeft = selfParent.GasPowerLeft
		prevMedianTime = selfParent.MedianTime
	} else if prevConfirmedHeader := p.PrevEpoch.LastHeaders[e.Creator]; prevConfirmedHeader != nil {
		prevGasPowerLeft = prevConfirmedHeader.GasPowerLeft
		if prevGasPowerLeft < startup {
			prevGasPowerLeft = startup
		}
		prevMedianTime = prevConfirmedHeader.MedianTime
	} else {
		prevGasPowerLeft = startup
		prevMedianTime = p.PrevEpoch.Time
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
