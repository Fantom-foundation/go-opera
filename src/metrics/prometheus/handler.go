package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/metrics"
)

const address = ":19090"

var log = logger.Get().WithField("module", "prometheus")

func init() {
	if !metrics.Enabled {
		return
	}

	enable()
}

func enable() {
	reg := metrics.DefaultRegistry
	reg.OnNew(collect)

	go func() {
		log.Infof("metrics server start on %s", address)
		defer log.Infof("metrics server is stopped")

		http.HandleFunc(
			"/metrics", promhttp.Handler().ServeHTTP)
		http.ListenAndServe(address, nil)

		// TODO: wait for exit signal?
	}()
}

func collect(name string, metric metrics.Metric) {
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
}
