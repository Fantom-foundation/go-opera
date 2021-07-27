package rate

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/metrics"
)

// Gauge represents an exponentially-weighted moving average of given values
type Gauge struct {
	input metrics.Meter
	count metrics.Meter
	max   int64
}

// NewGauge constructs a new Gauge and launches a goroutine.
// Be sure to call Stop() once the meter is of no use to allow for garbage collection.
func NewGauge() *Gauge {
	return &Gauge{
		input: metrics.NewMeterForced(),
		count: metrics.NewMeterForced(),
	}
}

// Mark records the the current value of the gauge.
func (g *Gauge) Mark(v int64) {
	g.input.Mark(v)
	g.count.Mark(1)
	//// maintain maximum input value in a thread-safe way
	for max := g.getMax(); max < v && !atomic.CompareAndSwapInt64(&g.max, max, v); {
		max = g.getMax()
	}
}

func (g *Gauge) getMax() int64 {
	return atomic.LoadInt64(&g.max)
}

func (g *Gauge) rateToGauge(valuesSum, calls float64) float64 {
	if calls < 0.000001 {
		return 0
	}
	gaugeValue := valuesSum / calls
	// gaugeValue cannot be larger than a maximum input value
	max := float64(g.getMax())
	if gaugeValue > max {
		return max
	}
	return gaugeValue
}

// Rate1 returns the one-minute moving average of the gauge values.
// Cannot be larger than max(largest input value, 0)
func (g *Gauge) Rate1() float64 {
	return g.rateToGauge(g.input.Rate1(), g.count.Rate1())
}

// Rate5 returns the five-minute moving average of the gauge values.
// Cannot be larger than max(largest input value, 0)
func (g *Gauge) Rate5() float64 {
	return g.rateToGauge(g.input.Rate5(), g.count.Rate5())
}

// Rate15 returns the fifteen-minute moving average of the gauge values.
// Cannot be larger than max(largest input value, 0)
func (g *Gauge) Rate15() float64 {
	return g.rateToGauge(g.input.Rate15(), g.count.Rate15())
}

// RateMean returns the gauge's mean value.
// Cannot be larger than max(largest input value, 0)
func (g *Gauge) RateMean() float64 {
	return g.rateToGauge(g.input.RateMean(), g.count.RateMean())
}

// Stop stops the gauge
func (g *Gauge) Stop() {
	g.input.Stop()
	g.count.Stop()
}
