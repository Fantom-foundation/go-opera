package prometheus

import (
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func convertToPrometheusMetric(name string, m interface{}) (prometheus.Collector, bool) {
	opts := prometheus.Opts{
		Namespace: "lachesis",
		Name:      prometheusDelims(name),
	}

	var collector prometheus.Collector

	switch metric := m.(type) {

	case metrics.Counter:
		collector = prometheus.NewCounterFunc(prometheus.CounterOpts(opts), func() float64 {
			return float64(metric.Count())
		})

	case metrics.Gauge:
		collector = prometheus.NewGaugeFunc(prometheus.GaugeOpts(opts), func() float64 {
			return float64(metric.Value())
		})

	case metrics.GaugeFloat64:
		collector = prometheus.NewGaugeFunc(prometheus.GaugeOpts(opts), func() float64 {
			return metric.Value()
		})

	// TODO: do
	case metrics.Meter:
		return nil, false
		/*m := metric.Snapshot()
		values["count"] = m.Count()
		values["1m.rate"] = m.Rate1()
		values["5m.rate"] = m.Rate5()
		values["15m.rate"] = m.Rate15()
		values["mean.rate"] = m.RateMean()*/

		// TODO: do
	case metrics.Healthcheck:
		return nil, false
	/*
		values["error"] = nil
		metric.Check()
		if err := metric.Error(); nil != err {
			values["error"] = metric.Error().Error()
		}
	*/

	// TODO: do
	case metrics.Histogram:
		return nil, false
	/*
		h := metric.Snapshot()
		ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
		values["count"] = h.Count()
		values["min"] = h.Min()
		values["max"] = h.Max()
		values["mean"] = h.Mean()
		values["stddev"] = h.StdDev()
		values["median"] = ps[0]
		values["75%"] = ps[1]
		values["95%"] = ps[2]
		values["99%"] = ps[3]
		values["99.9%"] = ps[4]
	*/

	// TODO: do
	case metrics.Timer:
		return nil, false
		/*
			t := metric.Snapshot()
			ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			values["count"] = t.Count()
			values["min"] = t.Min()
			values["max"] = t.Max()
			values["mean"] = t.Mean()
			values["stddev"] = t.StdDev()
			values["median"] = ps[0]
			values["75%"] = ps[1]
			values["95%"] = ps[2]
			values["99%"] = ps[3]
			values["99.9%"] = ps[4]
			values["1m.rate"] = t.Rate1()
			values["5m.rate"] = t.Rate5()
			values["15m.rate"] = t.Rate15()
			values["mean.rate"] = t.RateMean()
		*/

	default:
		logger.Warn("metric doesn't support prometheus",
			"metric", name,
			"type", reflect.TypeOf(m).String())
		return nil, false
	}

	return collector, true
}

func prometheusDelims(name string) string {
	return strings.ReplaceAll(name, "/", ":")
}
