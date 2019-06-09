package metrics

import (
	"sync"

	"github.com/pkg/errors"
)

// Registry management metrics
type Registry interface {
	// Register add new metric. If metric is exist, return error
	Register(name string, metric Metric) error

	// Unregister delete metric
	Unregister(name string)

	// UnregisterAll delete all metrics
	UnregisterAll()

	// Get the metric by name
	Get(name string) Metric

	// GetOrRegister getting existing metric or register new metric
	GetOrRegister(name string, metric Metric) Metric
}

var (
	// DefaultRegistry common in memory registry
	DefaultRegistry = NewRegistry()
)

// NewRegistry constructs a new Registry
func NewRegistry() Registry {
	return &inMemoryRegistry{metrics: new(sync.Map)}
}

type inMemoryRegistry struct {
	metrics *sync.Map
}

func (r *inMemoryRegistry) Register(name string, metric Metric) error {
	_, ok := r.metrics.Load(name)
	if ok {
		return errors.Errorf("metric '%s' is exist", name)
	}

	r.metrics.Store(name, metric)
	return nil
}

func (r *inMemoryRegistry) Unregister(name string) {
	r.metrics.Delete(name)
}

func (r *inMemoryRegistry) UnregisterAll() {
	r.metrics.Range(func(key, value interface{}) bool {
		r.Unregister(key.(string))
		return true
	})
}

func (r *inMemoryRegistry) Get(name string) Metric {
	value, ok := r.metrics.Load(name)
	if !ok {
		return nil
	}

	metric, ok := value.(Metric)
	if !ok {
		return nil
	}

	return metric
}

func (r *inMemoryRegistry) GetOrRegister(name string, metric Metric) Metric {
	existingMetric, ok := r.metrics.LoadOrStore(name, metric)
	if !ok {
		return metric
	}

	resultMetric, ok := existingMetric.(Metric)
	if !ok {
		return nil
	}

	return resultMetric
}
