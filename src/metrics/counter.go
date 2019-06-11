package metrics

import (
	"sync/atomic"
)

// Counter hold an int64 value that can be incremented and decremented.
type Counter interface {
	Metric

	// Inc add gain to current value
	Inc(int64)

	// Dec subtract gain from current value
	Dec(int64)

	// Reset current value to default value or 0
	Reset()

	// Value return current value
	Value() int64

	// Copy make snapshot
	Copy() Counter
}

// RegisterCounter create and register a new Counter.
func RegisterCounter(name string, r Registry) Counter {
	return RegisterPresetCounter(name, r, 0)
}

// RegisterPresetCounter create and register a new Counter with default value.
func RegisterPresetCounter(name string, r Registry, defaultValue int64) Counter {
	m := NewCounter(defaultValue)
	if r == nil {
		r = DefaultRegistry
	}

	r.Register(name, m)

	return m
}

// NewCounter constructs a new Counter.
func NewCounter(defaultValue int64) Counter {
	if !Enabled {
		return &nilCounter{
			Metric: newStandardMetric(nil),
		}
	}

	return &standardCounter{
		Metric:       newStandardMetric(nil),
		value:        defaultValue,
		defaultValue: defaultValue,
	}
}

type standardCounter struct {
	Metric

	value        int64
	defaultValue int64
}

func (c *standardCounter) Inc(i int64) {
	atomic.AddInt64(&c.value, i)
	c.Metric.updateModification()
}

func (c *standardCounter) Dec(i int64) {
	atomic.AddInt64(&c.value, -i)
	c.Metric.updateModification()
}

func (c *standardCounter) Reset() {
	atomic.StoreInt64(&c.value, c.defaultValue)
	c.Metric.updateModification()
}

func (c *standardCounter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

func (c *standardCounter) Copy() Counter {
	return &standardCounter{
		Metric:       c.Metric.copy(),
		value:        c.Value(),
		defaultValue: atomic.LoadInt64(&c.defaultValue),
	}
}

type nilCounter struct {
	Metric
}

func (c *nilCounter) Inc(int64) {
	c.Metric.updateModification()
}

func (c *nilCounter) Dec(int64) {
	c.Metric.updateModification()
}

func (c *nilCounter) Reset() {
	c.Metric.updateModification()
}

func (*nilCounter) Value() int64 {
	return 0
}

func (c *nilCounter) Copy() Counter {
	return &nilCounter{c.Metric.copy()}
}
