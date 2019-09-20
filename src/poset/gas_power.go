package poset

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
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

func (p *Poset) calcMemberGasPowerPerH(member common.Address) (perHour, maxStashed, startup uint64) {
	stake, ok := p.Members[member]
	if !ok {
		return 0, 0, 0
	}

	gas := p.dag.GasPower

	memberGasPowerPerH_bn := new(big.Int).SetUint64(gas.TotalPerH)
	mul(memberGasPowerPerH_bn, uint64(stake))
	div(memberGasPowerPerH_bn, uint64(p.Members.TotalStake()))
	perHour = memberGasPowerPerH_bn.Uint64()

	memberMaxStashed_bn := new(big.Int).Set(memberGasPowerPerH_bn)
	mul(memberMaxStashed_bn, uint64(gas.MaxStashedPeriod))
	div(memberMaxStashed_bn, uint64(time.Hour))
	maxStashed = memberMaxStashed_bn.Uint64()

	memberStartup_bn := new(big.Int).Set(memberGasPowerPerH_bn)
	mul(memberStartup_bn, uint64(gas.StartupPeriod))
	div(memberStartup_bn, uint64(time.Hour))
	startup = memberStartup_bn.Uint64()
	if startup < gas.MinStartupGasPower {
		startup = gas.MinStartupGasPower
	}

	return
}

func (p *Poset) CalcGasPower(e *inter.EventHeaderData) uint64 {
	gasPowerPerH, maxStashed, startup := p.calcMemberGasPowerPerH(e.Creator)

	var prevGasPowerLeft uint64
	var prevMedianTime inter.Timestamp

	if e.SelfParent() != nil {
		selfParent := p.GetEventHeader(e.Epoch, *e.SelfParent())
		prevGasPowerLeft = selfParent.GasPowerLeft
		prevMedianTime = selfParent.MedianTime
	} else if prevConfirmedHeader := p.PrevEpoch.LastHeaders[e.Creator]; prevConfirmedHeader != nil {
		prevGasPowerLeft = prevConfirmedHeader.GasPowerLeft
		prevMedianTime = prevConfirmedHeader.MedianTime
	} else {
		prevGasPowerLeft = startup
		prevMedianTime = p.PrevEpoch.Time
	}

	if prevGasPowerLeft > maxStashed {
		prevGasPowerLeft = maxStashed
	}
	if prevMedianTime > e.MedianTime {
		prevMedianTime = e.MedianTime // do not change e.MedianTime
	}

	gasPower_bn := new(big.Int).SetUint64(uint64(e.MedianTime))
	sub(gasPower_bn, uint64(prevMedianTime))
	mul(gasPower_bn, gasPowerPerH)
	div(gasPower_bn, uint64(time.Hour))

	return gasPower_bn.Uint64() + prevGasPowerLeft
}
