package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newStandardMetric(t *testing.T) {
	testFunc := func(t *testing.T, metric Metric, wandLoc *time.Location) {
		result := metric.(*standardMetric)
		assert.Equal(t, wandLoc, result.loc)
		assert.Equal(t, result.lastModification, result.creationTime)
		assert.Equal(t, result.loc, result.creationTime.Location())
	}

	t.Run("metrics turned on", func(t *testing.T) {
		Enabled = true
		loc := time.FixedZone("test", 123)
		metric := newStandardMetric(loc)
		testFunc(t, metric, loc)
	})

	t.Run("metrics turned off", func(t *testing.T) {
		Enabled = false
		metric := newStandardMetric(nil)
		assert.IsType(t, &nilMetric{}, metric)
	})

	t.Run("metrics without location", func(t *testing.T) {
		Enabled = true
		metric := newStandardMetric(nil)
		testFunc(t, metric, time.UTC)
	})
}

func Test_standardMetric_CreationTime(t *testing.T) {
	currentTime := time.Now()
	metric := &standardMetric{
		creationTime: currentTime,
	}

	assert.Equal(t, currentTime, metric.CreationTime())
}

func Test_standardMetric_LastModification(t *testing.T) {
	currentTime := time.Now()
	metric := &standardMetric{
		lastModification: currentTime,
	}

	assert.Equal(t, currentTime, metric.LastModification())
}

func Test_standardMetric_updateModification(t *testing.T) {
	loc := time.FixedZone("test", 123)
	currentTime := time.Now().In(loc)
	metric := &standardMetric{
		loc:              loc,
		lastModification: currentTime,
	}

	metric.updateModification()

	assert.True(t, metric.lastModification.UnixNano() > currentTime.UnixNano())
	assert.Equal(t, loc, metric.lastModification.Location())
}

func Test_standardMetric_copy(t *testing.T) {
	loc := time.FixedZone("test", 123)
	currentTime := time.Now().In(loc)
	updatedTime := currentTime.Add(5 * time.Second)
	metric := &standardMetric{
		loc:              loc,
		creationTime:     currentTime,
		lastModification: updatedTime,
	}
	Enabled = true

	copied := metric.copy()
	assert.False(t, copied == metric)
	assert.IsType(t, copied, &standardMetric{})

	result := copied.(*standardMetric)
	assert.Equal(t, metric.loc, result.loc)
	assert.True(t, result.creationTime.UnixNano() > metric.creationTime.UnixNano())
	assert.EqualValues(t, result.lastModification, result.creationTime)
}

func Test_nilMetric_CreationTime(t *testing.T) {
	metric := &nilMetric{}

	assert.Equal(t, int64(0), metric.CreationTime().UnixNano())
}

func Test_nilMetric_LastModification(t *testing.T) {
	metric := &nilMetric{}

	assert.Equal(t, int64(0), metric.LastModification().UnixNano())
}

func Test_nilMetric_copy(t *testing.T) {
	metric := &nilMetric{}

	copied := metric.copy()

	assert.False(t, copied == metric)
	assert.IsType(t, &nilMetric{}, copied)
}

func Test_nilMetric_updateModification(t *testing.T) {
	metric := &nilMetric{}

	assert.NotPanics(t, func() {
		metric.updateModification()
	})
}
