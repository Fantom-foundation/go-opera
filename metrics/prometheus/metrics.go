package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Fantom-foundation/go-lachesis/metrics"
)

func convertToPrometheusMetric(name string, metric metrics.Metric) (prometheus.Collector, bool) {
	opts := prometheus.Opts{
		Namespace: "lachesis",
		Name:      name,
	}

	var collector prometheus.Collector

	switch input := metric.(type) {
	case metrics.Counter:
		collector = prometheus.NewCounterFunc(prometheus.CounterOpts(opts), func() float64 {
			return float64(input.Value())
		})
	case metrics.Gauge:
		collector = prometheus.NewGaugeFunc(prometheus.GaugeOpts(opts), func() float64 {
			return float64(input.Value())
		})
	case metrics.GaugeFloat64:
		collector = prometheus.NewGaugeFunc(prometheus.GaugeOpts(opts), func() float64 {
			return input.Value()
		})
	default:
		return nil, false
	}

	return collector, true
}
