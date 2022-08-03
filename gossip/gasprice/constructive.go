package gasprice

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/utils/piecefunc"
)

func (gpo *Oracle) maxTotalGasPower() *big.Int {
	rules := gpo.backend.GetRules()

	allocBn := new(big.Int).SetUint64(rules.Economy.LongGasPower.AllocPerSec)
	periodBn := new(big.Int).SetUint64(uint64(rules.Economy.LongGasPower.MaxAllocPeriod))
	maxTotalGasPowerBn := new(big.Int).Mul(allocBn, periodBn)
	maxTotalGasPowerBn.Div(maxTotalGasPowerBn, secondBn)
	return maxTotalGasPowerBn
}

func (gpo *Oracle) effectiveMinGasPrice() *big.Int {
	return gpo.constructiveGasPrice(0, 0, gpo.backend.GetRules().Economy.MinGasPrice)
}

func (gpo *Oracle) constructiveGasPrice(gasOffestAbs uint64, gasOffestRatio uint64, adjustedMinPrice *big.Int) *big.Int {
	max := gpo.maxTotalGasPower()

	current64 := gpo.backend.TotalGasPowerLeft()
	if current64 > gasOffestAbs {
		current64 -= gasOffestAbs
	} else {
		current64 = 0
	}
	current := new(big.Int).SetUint64(current64)

	freeRatioBn := current.Mul(current, DecimalUnitBn)
	freeRatioBn.Div(freeRatioBn, max)
	freeRatio := freeRatioBn.Uint64()
	if freeRatio > gasOffestRatio {
		freeRatio -= gasOffestRatio
	} else {
		freeRatio = 0
	}
	if freeRatio > DecimalUnit {
		freeRatio = DecimalUnit
	}
	v := gpo.constructiveGasPriceOf(freeRatio, adjustedMinPrice)
	return v
}

var freeRatioToConstructiveGasRatio = piecefunc.NewFunc([]piecefunc.Dot{
	{
		X: 0,
		Y: 25 * DecimalUnit,
	},
	{
		X: 0.3 * DecimalUnit,
		Y: 9 * DecimalUnit,
	},
	{
		X: 0.5 * DecimalUnit,
		Y: 3.75 * DecimalUnit,
	},
	{
		X: 0.8 * DecimalUnit,
		Y: 1.5 * DecimalUnit,
	},
	{
		X: 0.95 * DecimalUnit,
		Y: 1.05 * DecimalUnit,
	},
	{
		X: DecimalUnit,
		Y: DecimalUnit,
	},
})

func (gpo *Oracle) constructiveGasPriceOf(freeRatio uint64, adjustedMinPrice *big.Int) *big.Int {
	multiplier := new(big.Int).SetUint64(freeRatioToConstructiveGasRatio(freeRatio))

	// gas price = multiplier * adjustedMinPrice
	price := multiplier.Mul(multiplier, adjustedMinPrice)
	return price.Div(price, DecimalUnitBn)
}
