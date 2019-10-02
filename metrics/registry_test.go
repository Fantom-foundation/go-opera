package metrics

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()

	assert.NotNil(t, reg)
	assert.IsType(t, &registry{}, reg)
	assert.NotNil(t, reg.(*registry))
}

func TestRegistryRegister(t *testing.T) {
	logger.SetTestMode(t)

	reg := newRegistry()
	name := "test"
	metric := newStandardMetric(nil)

	t.Run("without error", func(t *testing.T) {
		reg.Register(name, metric)

		loadedMetric, ok := reg.Load(name)

		assert.True(t, ok)
		assert.True(t, loadedMetric == metric)
	})

	t.Run("with error", func(t *testing.T) {
		reg.Store(name, metric)
		defer reg.Delete(name)

		err := reg.Register(name, metric)
		assert.Error(t, err)
	})
}

func TestRegistryGet(t *testing.T) {
	logger.SetTestMode(t)

	reg := newRegistry()
	name := "test"
	metric := newStandardMetric(nil)

	t.Run("not found metric", func(t *testing.T) {
		assert.Nil(t, reg.Get(name))
	})

	t.Run("metric is incorrect type", func(t *testing.T) {
		reg.Store(name, 0)
		assert.Nil(t, reg.Get(name))
	})

	t.Run("return a loaded metric", func(t *testing.T) {
		reg.Store(name, metric)

		result := reg.Get(name)

		assert.NotNil(t, result)
		assert.True(t, result == metric)
	})
}

func TestRegistryGetOrRegister(t *testing.T) {
	logger.SetTestMode(t)

	reg := newRegistry()
	name := "test"
	metric := newStandardMetric(nil)

	t.Run("new metric", func(t *testing.T) {
		result := reg.GetOrRegister(name, metric)
		defer reg.Delete(name)

		assert.NotNil(t, result)
		assert.True(t, result == metric)

		reg.Delete(name)
	})

	t.Run("existing metric is incorrect type", func(t *testing.T) {
		reg.Store(name, 0)
		defer reg.Delete(name)

		result := reg.GetOrRegister(name, metric)

		assert.Nil(t, result)
	})

	t.Run("return a loaded metric", func(t *testing.T) {
		existingMetric := &standardMetric{}
		reg.Store(name, existingMetric)
		defer reg.Delete(name)

		result := reg.GetOrRegister(name, metric)

		assert.NotNil(t, result)
		assert.False(t, result == metric)
		assert.True(t, result == existingMetric)
	})
}

func TestRegistryEach(t *testing.T) {
	logger.SetTestMode(t)

	reg := newRegistry()
	name := "test"
	metric := newStandardMetric(nil)
	noop := func() RegistryEachFunc { return func(name string, metric Metric) {} }

	t.Run("incorrect name type", func(t *testing.T) {
		noSupportName := 0
		reg.Store(noSupportName, metric)
		defer reg.Delete(noSupportName)

		err := reg.Each(noop())
		assert.Error(t, err)
	})

	t.Run("incorrect metric type", func(t *testing.T) {
		reg.Store(name, 0)
		defer reg.Delete(name)

		err := reg.Each(noop())
		assert.Error(t, err)
	})

	t.Run("correct types", func(t *testing.T) {
		reg.Store(name, metric)
		defer reg.Delete(name)

		counter := int32(0)
		reg.Each(func(name string, metric Metric) {
			atomic.AddInt32(&counter, 1)
		})

		assert.Equal(t, int32(1), counter)
	})

}

func TestRegistryOnNew(t *testing.T) {
	logger.SetTestMode(t)

	exp := []Metric{
		newStandardMetric(nil),
		newStandardMetric(nil),
	}

	var got []Metric
	subscriber := func(name string, metric Metric) {
		got = append(got, metric)
	}

	reg := newRegistry()
	reg.Register("before", exp[0])
	reg.OnNew(subscriber)
	_ = reg.GetOrRegister("after", exp[1])

	assert.Equal(t, exp, got)
}
