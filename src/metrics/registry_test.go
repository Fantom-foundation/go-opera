package metrics

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package=metrics -source=registry.go -destination=registry_mock_test.go Registry

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	assert.NotNil(t, registry)
	assert.IsType(t, &inMemoryRegistry{}, registry)
	assert.NotNil(t, registry.(*inMemoryRegistry).metrics)
}

func Test_registry_Register(t *testing.T) {
	registry := &inMemoryRegistry{
		metrics: new(sync.Map),
	}
	name := "test"
	metric := newStandardMetric(nil)

	t.Run("without error", func(t *testing.T) {
		registry.Register(name, metric)

		loadedMetric, ok := registry.metrics.Load(name)

		assert.True(t, ok)
		assert.True(t, loadedMetric == metric)
	})

	t.Run("with error", func(t *testing.T) {
		registry.metrics.Store(name, metric)
		defer registry.metrics.Delete(name)

		log.Logger.ExitFunc = func(i int) {
			assert.Equal(t, 1, i)
		}
		defer func() {
			log.Logger.ExitFunc = nil
		}()

		registry.Register(name, metric)
	})
}

func Test_registry_Unregister(t *testing.T) {
	registry := &inMemoryRegistry{
		metrics: new(sync.Map),
	}
	name := "test"
	metric := newStandardMetric(nil)

	assert.NotPanics(t, func() {
		registry.metrics.Store(name, metric)
		registry.Unregister(name)
	})

	_, ok := registry.metrics.Load(name)
	assert.False(t, ok)
}

func Test_registry_UnregisterAll(t *testing.T) {
	registry := &inMemoryRegistry{
		metrics: new(sync.Map),
	}
	names := []string{"0", "1", "2"}
	metric := newStandardMetric(nil)

	assert.NotPanics(t, func() {
		for _, name := range names {
			registry.metrics.Store(name, metric)
		}

		registry.UnregisterAll()
	})

	registry.metrics.Range(func(key, value interface{}) bool {
		assert.Fail(t, "some metric have registered")
		return true
	})
}

func Test_registry_Get(t *testing.T) {
	registry := &inMemoryRegistry{
		metrics: new(sync.Map),
	}
	name := "test"
	metric := newStandardMetric(nil)

	t.Run("not found metric", func(t *testing.T) {
		assert.Nil(t, registry.Get(name))
	})

	t.Run("metric is incorrect type", func(t *testing.T) {
		registry.metrics.Store(name, 0)
		assert.Nil(t, registry.Get(name))
	})

	t.Run("return a loaded metric", func(t *testing.T) {
		registry.metrics.Store(name, metric)

		result := registry.Get(name)

		assert.NotNil(t, result)
		assert.True(t, result == metric)
	})
}

func Test_registry_GetOrRegister(t *testing.T) {
	registry := &inMemoryRegistry{
		metrics: new(sync.Map),
	}
	name := "test"
	metric := newStandardMetric(nil)

	t.Run("new metric", func(t *testing.T) {
		result := registry.GetOrRegister(name, metric)
		defer registry.metrics.Delete(name)

		assert.NotNil(t, result)
		assert.True(t, result == metric)

		registry.metrics.Delete(name)
	})

	t.Run("existing metric is incorrect type", func(t *testing.T) {
		registry.metrics.Store(name, 0)
		defer registry.metrics.Delete(name)

		result := registry.GetOrRegister(name, metric)

		assert.Nil(t, result)
	})

	t.Run("return a loaded metric", func(t *testing.T) {
		existingMetric := &standardMetric{}
		registry.metrics.Store(name, existingMetric)
		defer registry.metrics.Delete(name)

		result := registry.GetOrRegister(name, metric)

		assert.NotNil(t, result)
		assert.False(t, result == metric)
		assert.True(t, result == existingMetric)
	})
}

func Test_inMemoryRegistry_Each(t *testing.T) {
	registryFunc := func() RegistryEachFunc { return func(name string, metric Metric) {} }
	registry := &inMemoryRegistry{
		metrics: new(sync.Map),
	}
	name := "test"
	metric := newStandardMetric(nil)

	t.Run("incorrect name type", func(t *testing.T) {
		noSupportName := 0
		registry.metrics.Store(noSupportName, metric)
		defer registry.metrics.Delete(noSupportName)

		log.Logger.ExitFunc = func(i int) {
			assert.Equal(t, 1, i)
		}
		defer func() {
			log.Logger.ExitFunc = nil
		}()

		registry.Each(registryFunc())
	})

	t.Run("incorrect metric type", func(t *testing.T) {
		registry.metrics.Store(name, 0)
		defer registry.metrics.Delete(name)

		log.Logger.ExitFunc = func(i int) {
			assert.Equal(t, 1, i)
		}
		defer func() {
			log.Logger.ExitFunc = nil
		}()

		registry.Each(registryFunc())
	})

	t.Run("correct types", func(t *testing.T) {
		registry.metrics.Store(name, metric)
		defer registry.metrics.Delete(name)

		counter := int32(0)
		registry.Each(func(name string, metric Metric) {
			atomic.AddInt32(&counter, 1)
		})

		assert.Equal(t, int32(1), counter)
	})

}
