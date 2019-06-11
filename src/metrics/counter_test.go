package metrics

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterCounter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name := "test"
	registry := NewMockRegistry(ctrl)

	t.Run("registry is set", func(t *testing.T) {
		registry.EXPECT().Register(name, gomock.Any()).Return()

		resultCounter := RegisterCounter(name, registry)

		assert.NotNil(t, resultCounter)
	})

	t.Run("registry isn't set", func(t *testing.T) {
		DefaultRegistry = registry
		registry.EXPECT().Register(name, gomock.Any()).Return()

		counter := RegisterCounter(name, nil)

		assert.NotNil(t, counter)
	})
}

func TestRegisterPresetCounter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name := "test"

	registry := NewMockRegistry(ctrl)

	t.Run("registry is set", func(t *testing.T) {
		registry.EXPECT().Register(name, gomock.Any()).Return()

		counter := RegisterPresetCounter(name, registry, 0)

		assert.NotNil(t, counter)
	})

	t.Run("is set default value", func(t *testing.T) {
		defaultValue := int64(123)
		Enabled = true
		registry.EXPECT().Register(name, gomock.Any()).Return()

		resultCounter := RegisterPresetCounter(name, registry, defaultValue)

		assert.NotNil(t, resultCounter)
		assert.IsType(t, &standardCounter{}, resultCounter)
		assert.Equal(t, defaultValue, resultCounter.(*standardCounter).defaultValue)
	})

	t.Run("registry isn't set", func(t *testing.T) {
		DefaultRegistry = registry
		registry.EXPECT().Register(name, gomock.Any()).Return()

		counter := RegisterPresetCounter(name, nil, 0)

		assert.NotNil(t, counter)
	})
}

func TestNewCounter(t *testing.T) {
	t.Run("metrics turn on", func(t *testing.T) {
		Enabled = true
		defaultValue := int64(123)

		result := NewCounter(defaultValue)

		assert.NotNil(t, result)
		assert.IsType(t, &standardCounter{}, result)

		resultWithType := result.(*standardCounter)
		assert.NotNil(t, resultWithType.Metric)
		assert.Equal(t, defaultValue, resultWithType.value)
		assert.Equal(t, defaultValue, resultWithType.defaultValue)
	})

	t.Run("metrics turn off", func(t *testing.T) {
		Enabled = false

		result := NewCounter(0)

		assert.NotNil(t, result)
		assert.IsType(t, &nilCounter{}, result)

		resultWithType := result.(*nilCounter)
		assert.NotNil(t, resultWithType.Metric)
	})
}

func Test_counter_Inc(t *testing.T) {
	defaultValue := int64(1)
	counter := &standardCounter{
		Metric:       &nilMetric{},
		value:        defaultValue,
		defaultValue: defaultValue,
	}
	step := int64(2)

	counter.Inc(step)

	assert.NotNil(t, counter.Metric)
	assert.Equal(t, defaultValue+step, counter.value)
	assert.Equal(t, defaultValue, counter.defaultValue)
}

func Test_counter_Dec(t *testing.T) {
	defaultValue := int64(1)
	counter := &standardCounter{
		Metric:       &nilMetric{},
		value:        defaultValue,
		defaultValue: defaultValue,
	}
	step := int64(2)

	counter.Dec(step)

	assert.NotNil(t, counter.Metric)
	assert.Equal(t, defaultValue-step, counter.value)
	assert.Equal(t, defaultValue, counter.defaultValue)
}

func Test_counter_Reset(t *testing.T) {
	defaultValue := int64(1)
	counter := &standardCounter{
		Metric:       &nilMetric{},
		value:        defaultValue,
		defaultValue: defaultValue,
	}

	counter.Reset()

	assert.NotNil(t, counter.Metric)
	assert.Equal(t, defaultValue, counter.defaultValue)
	assert.Equal(t, counter.defaultValue, counter.value)
}

func Test_counter_Value(t *testing.T) {
	defaultValue := int64(1)
	value := int64(100)
	counter := &standardCounter{
		Metric:       &nilMetric{},
		value:        value,
		defaultValue: defaultValue,
	}

	result := counter.Value()

	assert.NotNil(t, counter.Metric)
	assert.Equal(t, defaultValue, counter.defaultValue)
	assert.Equal(t, value, result)
}

func Test_counter_Copy(t *testing.T) {
	defaultValue := int64(1)
	value := int64(100)
	counter := &standardCounter{
		Metric:       &nilMetric{},
		value:        value,
		defaultValue: defaultValue,
	}

	copied := counter.Copy()

	assert.False(t, copied == counter)
	assert.IsType(t, &standardCounter{}, copied)
}

func Test_nilCounter_Inc(t *testing.T) {
	counter := &nilCounter{
		Metric: &nilMetric{},
	}

	assert.NotPanics(t, func() {
		counter.Inc(1)
	})
}

func Test_nilCounter_Dec(t *testing.T) {
	counter := &nilCounter{
		Metric: &nilMetric{},
	}

	assert.NotPanics(t, func() {
		counter.Dec(1)
	})
}

func Test_nilCounter_Reset(t *testing.T) {
	counter := &nilCounter{
		Metric: &nilMetric{},
	}

	assert.NotPanics(t, func() {
		counter.Reset()
	})
}

func Test_nilCounter_Value(t *testing.T) {
	counter := &nilCounter{
		Metric: &nilMetric{},
	}

	assert.Equal(t, int64(0), counter.Value())
}

func Test_nilCounter_Copy(t *testing.T) {
	counter := &nilCounter{
		Metric: &nilMetric{},
	}

	copied := counter.Copy()

	assert.False(t, copied == counter)
	assert.IsType(t, &nilCounter{}, copied)
}
