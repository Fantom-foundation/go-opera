package rate

import "github.com/ethereum/go-ethereum/metrics"

// Gauge represents an exponentially-weighted moving average of given values
type Gauge struct {
	input metrics.Meter
	all   metrics.Meter
}

// NewGauge constructs a new Gauge and launches a goroutine.
// Be sure to call Stop() once the meter is of no use to allow for garbage collection.
func NewGauge() *Gauge {
	return &Gauge{
		input: metrics.NewMeterForced(),
		all:   metrics.NewMeterForced(),
	}
}

// Mark records the the current value of the gauge.
func (g *Gauge) Mark(v int64) {
	g.input.Mark(v)
	g.all.Mark(1)
}

// Rate1 returns the one-minute moving average of the gauge values.
func (g *Gauge) Rate1() float64 {
	allRate := g.all.Rate1()
	if allRate < 0.01 {
		return 0
	}
	return g.input.Rate1() / allRate
}

// Rate1 returns the five-minute moving average of the gauge values.
func (g *Gauge) Rate5() float64 {
	allRate := g.all.Rate5()
	if allRate < 0.001 {
		return 0
	}
	return g.input.Rate5() / allRate
}

// Rate1 returns the fifteen-minute moving average of the gauge values.
func (g *Gauge) Rate15() float64 {
	allRate := g.all.Rate15()
	if allRate < 0.0001 {
		return 0
	}
	return g.input.Rate15() / allRate
}

// RateMean returns the gauge's mean value.
func (g *Gauge) RateMean() float64 {
	allRate := g.all.RateMean()
	if allRate < 0.00001 {
		return 0
	}
	return g.input.RateMean() / allRate
}

// Stop stops the gauge
func (g *Gauge) Stop() {
	g.input.Stop()
	g.all.Stop()
}
