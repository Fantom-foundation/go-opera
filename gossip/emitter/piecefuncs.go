package emitter

import (
	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/piecefunc"
)

var (
	// confirmingEmitIntervalF is a piecewise function for validator confirming internal depending on a stake amount before him
	confirmingEmitIntervalF = piecefunc.NewFunc([]piecefunc.Dot{
		{
			X: 0,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
		{
			X: 0.78 * piecefunc.DecimalUnit,
			Y: 1.1 * piecefunc.DecimalUnit,
		},
		{
			X: 0.8 * piecefunc.DecimalUnit,
			Y: 10.0 * piecefunc.DecimalUnit,
		},
		{ // validators >0.8 emit confirming events very rarely
			X: 0.81 * piecefunc.DecimalUnit,
			Y: 50.0 * piecefunc.DecimalUnit,
		},
		{ // validators >0.8 emit confirming events very rarely
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 60.0 * piecefunc.DecimalUnit,
		},
	})
	// scalarUpdMetricF is a piecewise function for validator's event metric diff depending on a number of newly observed events
	scalarUpdMetricF = piecefunc.NewFunc([]piecefunc.Dot{
		{
			X: 0,
			Y: 0,
		},
		{ // first observed event gives a major metric diff
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 0.66 * piecefunc.DecimalUnit,
		},
		{ // second observed event gives a minor diff
			X: 2.0 * piecefunc.DecimalUnit,
			Y: 0.8 * piecefunc.DecimalUnit,
		},
		{ // other observed event give only a subtle diff
			X: 8.0 * piecefunc.DecimalUnit,
			Y: 0.99 * piecefunc.DecimalUnit,
		},
		{
			X: 100.0 * piecefunc.DecimalUnit,
			Y: 0.999 * piecefunc.DecimalUnit,
		},
		{
			X: 10000.0 * piecefunc.DecimalUnit,
			Y: 0.9999 * piecefunc.DecimalUnit,
		},
	})
	// eventMetricF is a piecewise function for event metric adjustment depending on a non-adjusted event metric
	eventMetricF = piecefunc.NewFunc([]piecefunc.Dot{
		{ // event metric is never zero
			X: 0,
			Y: 0.005 * piecefunc.DecimalUnit,
		},
		{
			X: 0.01 * piecefunc.DecimalUnit,
			Y: 0.03 * piecefunc.DecimalUnit,
		},
		{ // if metric is below ~0.2, then validator shouldn't emit event unless waited very long
			X: 0.2 * piecefunc.DecimalUnit,
			Y: 0.05 * piecefunc.DecimalUnit,
		},
		{
			X: 0.3 * piecefunc.DecimalUnit,
			Y: 0.22 * piecefunc.DecimalUnit,
		},
		{ // ~0.3-0.5 is an optimal metric to emit an event
			X: 0.4 * piecefunc.DecimalUnit,
			Y: 0.45 * piecefunc.DecimalUnit,
		},
		{
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
	})
	validatorsToOverheadF = piecefunc.NewFunc([]piecefunc.Dot{
		{
			X: 0,
			Y: 0,
		},
		{
			X: 25,
			Y: 0.05 * piecefunc.DecimalUnit,
		},
		{
			X: 50,
			Y: 0.1 * piecefunc.DecimalUnit,
		},
		{
			X: 100,
			Y: 0.4 * piecefunc.DecimalUnit,
		},
		{
			X: 200,
			Y: 0.9 * piecefunc.DecimalUnit,
		},
		{
			X: 1000,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
	})
	overheadF = func(validatorsNum idx.Validator, busyRate uint64) uint64 {
		if busyRate > piecefunc.DecimalUnit {
			busyRate = piecefunc.DecimalUnit
		}
		return validatorsToOverheadF(uint64(validatorsNum)) * busyRate / piecefunc.DecimalUnit
	}
	overheadAdjustedEventMetricF = func(validatorsNum idx.Validator, busyRate uint64, eventMetric ancestor.Metric) ancestor.Metric {
		return ancestor.Metric(piecefunc.DecimalUnit-overheadF(validatorsNum, busyRate)) * eventMetric / piecefunc.DecimalUnit
	}
)
