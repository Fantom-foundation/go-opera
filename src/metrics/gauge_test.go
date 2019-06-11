package metrics

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterGauge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name := "test"
	registry := NewMockRegistry(ctrl)

	t.Run("registry is set", func(t *testing.T) {
		registry.EXPECT().Register(name, gomock.Any()).Return()

		gauge := RegisterGauge(name, registry)

		assert.NotNil(t, gauge)
	})

	t.Run("registry isn't set", func(t *testing.T) {
		DefaultRegistry = registry
		registry.EXPECT().Register(name, gomock.Any()).Return()

		gauge := RegisterGauge(name, nil)

		assert.NotNil(t, gauge)
	})
}

func TestNewGauge(t *testing.T) {
	t.Run("metrics turn on", func(t *testing.T) {
		Enabled = true

		gauge := NewGauge()

		assert.NotNil(t, gauge)
		assert.IsType(t, &standardGauge{}, gauge)
	})

	t.Run("metrics turn off", func(t *testing.T) {
		Enabled = false

		gauge := NewGauge()

		assert.NotNil(t, gauge)
		assert.IsType(t, &nilGauge{}, gauge)
	})
}

func Test_standardGauge_Update(t *testing.T) {
	gauge := &standardGauge{
		Metric: &nilMetric{},
		value:  0,
	}
	newValue := int64(123)

	assert.NotPanics(t, func() {
		gauge.Update(newValue)
	})

	assert.Equal(t, newValue, gauge.value)
}

func Test_standardGauge_Value(t *testing.T) {
	value := int64(123)
	gauge := &standardGauge{
		Metric: &nilMetric{},
		value:  value,
	}

	assert.NotPanics(t, func() {
		assert.Equal(t, value, gauge.Value())
	})
}

func Test_standardGauge_Copy(t *testing.T) {
	value := int64(123)
	gauge := &standardGauge{
		Metric: &nilMetric{},
		value:  value,
	}

	copied := gauge.Copy()

	assert.NotNil(t, copied)
	assert.IsType(t, &standardGauge{}, copied)
	assert.False(t, copied == gauge)

	result := copied.(*standardGauge)
	assert.Equal(t, value, result.value)
}

func Test_nilGauge_Update(t *testing.T) {
	gauge := &nilGauge{
		Metric: &nilMetric{},
	}

	assert.NotPanics(t, func() {
		gauge.Update(1)
	})
}

func Test_nilGauge_Value(t *testing.T) {
	gauge := &nilGauge{
		Metric: &nilGauge{},
	}

	assert.NotPanics(t, func() {
		value := gauge.Value()

		assert.Equal(t, int64(0), value)
	})
}

func Test_nilGauge_Copy(t *testing.T) {
	gauge := &nilGauge{
		Metric: &nilMetric{},
	}

	copied := gauge.Copy()

	assert.NotNil(t, copied)
	assert.IsType(t, &nilGauge{}, copied)
	assert.False(t, copied == gauge)
}
