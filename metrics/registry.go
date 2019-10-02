//go:generate mockgen -package=metrics -self_package=github.com/Fantom-foundation/go-lachesis/src/metrics -destination=registry_mock_test.go github.com/Fantom-foundation/go-lachesis/src/metrics Registry

package metrics

import (
	"fmt"
	"sync"
)

// RegistryEachFunc type for call back function Registry.Each.
type RegistryEachFunc func(name string, metric Metric)

// Registry management metrics.
type Registry interface {
	// Each iteration on all registered metrics.
	Each(f RegistryEachFunc) error

	// Register add new metric. If metric is exist, write Fatal error.
	Register(name string, metric Metric) error

	// Get the metric by name.
	Get(name string) Metric

	// GetOrRegister getting existing metric or register new metric.
	GetOrRegister(name string, metric Metric) Metric

	// OnNew subscribes f on new Metric registration.
	OnNew(f RegistryEachFunc)
}

var (
	// DefaultRegistry common in memory registry.
	DefaultRegistry = NewRegistry()
)

// NewRegistry constructs a new Registry.
func NewRegistry() Registry {
	return newRegistry()
}

type registry struct {
	*sync.Map
	subscribers []RegistryEachFunc
}

func newRegistry() *registry {
	return &registry{
		new(sync.Map), nil,
	}
}

func (r *registry) Each(f RegistryEachFunc) (err error) {
	r.Range(func(key, value interface{}) bool {
		name, ok := key.(string)
		if !ok {
			err = fmt.Errorf("key name must be string")
			return false
		}

		metric, ok := value.(Metric)
		if !ok {
			err = fmt.Errorf("metric must be instance of Metric")
			return false
		}

		f(name, metric)

		return true
	})

	return err
}

func (r *registry) Register(name string, metric Metric) error {
	_, ok := r.Load(name)
	if ok {
		return fmt.Errorf("metric name '%s' is already registered", name)
	}

	r.Store(name, metric)
	r.onNew(name, metric)

	return nil
}

func (r *registry) Get(name string) Metric {
	value, ok := r.Load(name)
	if !ok {
		return nil
	}

	metric, ok := value.(Metric)
	if !ok {
		return nil
	}

	return metric
}

func (r *registry) GetOrRegister(name string, metric Metric) Metric {
	existingMetric, ok := r.LoadOrStore(name, metric)
	if !ok {
		r.onNew(name, metric)
		return metric
	}

	resultMetric, ok := existingMetric.(Metric)
	if !ok {
		return nil
	}

	return resultMetric
}

func (r *registry) OnNew(f RegistryEachFunc) {
	r.subscribers = append(r.subscribers, f)
	r.Each(f)
}

func (r *registry) onNew(name string, metric Metric) {
	for _, f := range r.subscribers {
		f(name, metric)
	}
}
