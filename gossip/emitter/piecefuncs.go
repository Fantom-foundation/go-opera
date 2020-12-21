package emitter

import (
	"math"

	"github.com/Fantom-foundation/go-opera/gossip/emitter/piecefunc"
)

var (
	confirmingEmitIntervalF = []piecefunc.Dot{
		{
			X: 0,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
		{
			X: 0.33 * piecefunc.DecimalUnit,
			Y: 1.05 * piecefunc.DecimalUnit,
		},
		{
			X: 0.66 * piecefunc.DecimalUnit,
			Y: 1.20 * piecefunc.DecimalUnit,
		},
		{
			X: 0.8 * piecefunc.DecimalUnit,
			Y: 1.5 * piecefunc.DecimalUnit,
		},
		{
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 1000.0 * piecefunc.DecimalUnit,
		},
	}
	scalarUpdMetricF = []piecefunc.Dot{
		{
			X: 0,
			Y: 0,
		},
		{
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 0.66 * piecefunc.DecimalUnit,
		},
		{
			X: 2.0 * piecefunc.DecimalUnit,
			Y: 0.8 * piecefunc.DecimalUnit,
		},
		{
			X: 8.0 * piecefunc.DecimalUnit,
			Y: 0.99 * piecefunc.DecimalUnit,
		},
		{
			X: math.MaxUint32 * piecefunc.DecimalUnit,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
	}
	eventMetricF = []piecefunc.Dot{
		{
			X: 0,
			Y: 0.04 * piecefunc.DecimalUnit,
		},
		{
			X: 0.2 * piecefunc.DecimalUnit,
			Y: 0.05 * piecefunc.DecimalUnit,
		},
		{
			X: 0.3 * piecefunc.DecimalUnit,
			Y: 0.22 * piecefunc.DecimalUnit,
		},
		{
			X: 0.4 * piecefunc.DecimalUnit,
			Y: 0.45 * piecefunc.DecimalUnit,
		},
		{
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
		{
			X: math.MaxUint32 * piecefunc.DecimalUnit,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
	}
)
