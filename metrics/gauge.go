package metrics

import (
	"sync/atomic"
)

// Gauge hold an int64 value.
type Gauge interface {
	Metric

	// Update change property
	Update(int64)

	// Value return current property
	Value() int64

	// Copy make snapshot
	Copy() Gauge
}

// RegisterGauge create and register a new Gauge.
func RegisterGauge(name string, r Registry) Gauge {
	m := NewGauge()
	if r == nil {
		r = DefaultRegistry
	}

	r.Register(name, m)

	return m
}

// NewGauge constructs a new Gauge.
func NewGauge() Gauge {
	if !Enabled {
		return &nilGauge{
			Metric: newStandardMetric(nil),
		}
	}

	return &standardGauge{
		Metric: newStandardMetric(nil),
		value:  0,
	}
}

type standardGauge struct {
	Metric

	value int64
}

func (g *standardGauge) Update(v int64) {
	atomic.StoreInt64(&g.value, v)
	g.Metric.updateModification()
}

func (g *standardGauge) Value() int64 {
	return atomic.LoadInt64(&g.value)
}

func (g *standardGauge) Copy() Gauge {
	return &standardGauge{
		Metric: g.Metric.copy(),
		value:  g.Value(),
	}
}

type nilGauge struct {
	Metric
}

func (g *nilGauge) Update(int64) {
	g.Metric.updateModification()
}

func (*nilGauge) Value() int64 {
	return 0
}

func (g *nilGauge) Copy() Gauge {
	return &nilGauge{
		Metric: g.Metric.copy(),
	}
}
