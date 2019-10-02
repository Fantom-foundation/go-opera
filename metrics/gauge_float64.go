package metrics

import (
	"sync"
)

// GaugeFloat64 hold a float64 value.
type GaugeFloat64 interface {
	Metric

	// Update change property
	Update(float64)

	// Value return current property
	Value() float64

	// Copy make snapshot
	Copy() GaugeFloat64
}

// RegisterGaugeFloat64 create and register a new GaugeFloat64.
func RegisterGaugeFloat64(name string, r Registry) GaugeFloat64 {
	m := NewGaugeFloat64()
	if r == nil {
		r = DefaultRegistry
	}

	r.Register(name, m)

	return m
}

// NewGaugeFloat64 constructs a new GaugeFloat64.
func NewGaugeFloat64() GaugeFloat64 {
	if !Enabled {
		return &nilGaugeFloat64{
			Metric: newStandardMetric(nil),
		}
	}

	return &standardGaugeFloat64{
		Metric: newStandardMetric(nil),
		value:  0.0,
	}
}

type standardGaugeFloat64 struct {
	Metric

	mu    sync.RWMutex
	value float64
}

func (g *standardGaugeFloat64) Update(v float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.value = v
	g.Metric.updateModification()
}

func (g *standardGaugeFloat64) Value() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.value
}

func (g *standardGaugeFloat64) Copy() GaugeFloat64 {
	return &standardGaugeFloat64{
		Metric: g.Metric.copy(),
		value:  g.Value(),
	}
}

type nilGaugeFloat64 struct {
	Metric
}

func (g *nilGaugeFloat64) Update(float64) {
	g.Metric.updateModification()
}

func (*nilGaugeFloat64) Value() float64 {
	return 0.0
}

func (g *nilGaugeFloat64) Copy() GaugeFloat64 {
	return &nilGaugeFloat64{
		Metric: g.Metric.copy(),
	}
}
