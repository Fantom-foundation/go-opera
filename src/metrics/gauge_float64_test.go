package metrics

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterGaugeFloat64(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name := "test"
	registry := NewMockRegistry(ctrl)

	t.Run("registry is set", func(t *testing.T) {
		registry.EXPECT().Register(name, gomock.Any()).Return()

		gauge := RegisterGaugeFloat64(name, registry)

		assert.NotNil(t, gauge)
	})

	t.Run("registry isn't set", func(t *testing.T) {
		DefaultRegistry = registry
		registry.EXPECT().Register(name, gomock.Any()).Return()

		gauge := RegisterGaugeFloat64(name, nil)

		assert.NotNil(t, gauge)
	})
}

func TestNewGaugeFloat64(t *testing.T) {
	t.Run("metrics turn on", func(t *testing.T) {
		Enabled = true

		gauge := NewGaugeFloat64()

		assert.NotNil(t, gauge)
		assert.IsType(t, &standardGaugeFloat64{}, gauge)
	})

	t.Run("metrics turn off", func(t *testing.T) {
		Enabled = false

		gauge := NewGaugeFloat64()

		assert.NotNil(t, gauge)
		assert.IsType(t, &nilGaugeFloat64{}, gauge)
	})
}

func Test_standardGaugeFloat64_Update(t *testing.T) {
	gauge := &standardGaugeFloat64{
		Metric: &nilMetric{},
	}
	value := float64(1.2)

	gauge.Update(value)

	assert.Equal(t, value, gauge.value)
}

func Test_standardGaugeFloat64_Value(t *testing.T) {
	value := float64(1.2)
	gauge := &standardGaugeFloat64{
		Metric: &nilMetric{},
		value:  value,
	}

	assert.Equal(t, value, gauge.Value())
}

func Test_standardGaugeFloat64_Copy(t *testing.T) {
	value := float64(1.2)
	gauge := &standardGaugeFloat64{
		Metric: &nilMetric{},
		value:  value,
	}

	copied := gauge.Copy()

	assert.NotNil(t, copied)
	assert.IsType(t, &standardGaugeFloat64{}, copied)
	assert.False(t, copied == gauge)

	result := copied.(*standardGaugeFloat64)
	assert.Equal(t, value, result.value)
}

func Test_nilGaugeFloat64_Update(t *testing.T) {
	gauge := &nilGaugeFloat64{
		Metric: &nilMetric{},
	}

	assert.NotPanics(t, func() {
		gauge.Update(1.1)
	})
}

func Test_nilGaugeFloat64_Value(t *testing.T) {
	gauge := &nilGaugeFloat64{
		Metric: &nilMetric{},
	}

	assert.Equal(t, float64(0.0), gauge.Value())
}

func Test_nilGaugeFloat64_Copy(t *testing.T) {
	gauge := &nilGaugeFloat64{
		Metric: &nilMetric{},
	}

	copied := gauge.Copy()

	assert.NotNil(t, copied)
	assert.IsType(t, &nilGaugeFloat64{}, copied)
	assert.False(t, copied == gauge)
}
