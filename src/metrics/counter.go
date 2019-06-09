package metrics

import (
	"sync/atomic"
)

type Counter interface {
	Metric

	Inc(int64)
	Dec(int64)
	Reset()
	Value() int64
	Copy() Counter
}

func NewRegisteredCounter(name string, r Registry) (Counter, error) {
	return NewRegisteredPresetCounter(name, r, 0)
}

func NewRegisteredPresetCounter(name string, r Registry, defaultValue int64) (Counter, error) {
	c := NewCounter(defaultValue)
	if r == nil {
		r = DefaultRegistry
	}

	err := r.Register(name, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func NewCounter(defaultValue int64) Counter {
	if !Enabled {
		return &nilCounter{
			Metric: &nilMetric{},
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
	c.updateModification()
}

func (c *standardCounter) Dec(i int64) {
	atomic.AddInt64(&c.value, -i)
	c.updateModification()
}

func (c *standardCounter) Reset() {
	atomic.StoreInt64(&c.value, c.defaultValue)
	c.updateModification()
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

func (*nilCounter) Inc(int64) {}

func (*nilCounter) Dec(int64) {}

func (*nilCounter) Reset() {}

func (*nilCounter) Value() int64 {
	return 0
}

func (c *nilCounter) Copy() Counter {
	return &nilCounter{c.Metric.copy()}
}
