package metrics

import (
	"sync"
)

type RegistryEachFunc func(name string, metric Metric)

// Registry management metrics
type Registry interface {
	Each(f RegistryEachFunc)

	// Register add new metric. If metric is exist, write Fatal error
	Register(name string, metric Metric)

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

func (r *inMemoryRegistry) Each(f RegistryEachFunc) {
	r.metrics.Range(func(key, value interface{}) bool {
		name, ok := key.(string)
		if !ok {
			log.Fatal("name must be string")
		}

		metric, ok := value.(Metric)
		if !ok {
			log.Fatal("metric is incorrect type: must be Metric type")
		}

		f(name, metric)

		return true
	})
}

func (r *inMemoryRegistry) Register(name string, metric Metric) {
	_, ok := r.metrics.Load(name)
	if ok {
		log.Fatalf("metric '%s' is exist", name)
	}

	r.metrics.Store(name, metric)
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
