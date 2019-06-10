package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/metrics"
)

var log = logger.Get().WithField("module", "prometheus")

func Handler(registry metrics.Registry) http.Handler {
	handler := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer handler.ServeHTTP(w, r)

		registry.Each(func(name string, metric metrics.Metric) {
			collector, ok := convertToPrometheusMetric(name, metric)
			if !ok {
				log.Debugf("metric '%s' not support prometheus", name)
				return
			}

			err := prometheus.Register(collector)
			switch err.(type) {
			case prometheus.AlreadyRegisteredError:
				return
			default:
			}
			if err != nil {
				log.Error(err)
			}
		})

	})
}
